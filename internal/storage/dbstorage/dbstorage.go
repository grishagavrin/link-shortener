// Package dbstorage contains methods for postgreSQL storage work
package dbstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct {
	dbi     *pgxpool.Pool
	l       *zap.Logger
	chBatch chan models.BatchDelete
}

// New initialize new table in postgreSQL storage
func New(dbi *pgxpool.Pool, l *zap.Logger, ch chan models.BatchDelete) (*PostgreSQLStorage, error) {
	// Check if scheme exist
	sql := `
	CREATE TABLE IF NOT EXISTS public.short_links(
		id serial,
		user_id varchar(50),
		origin  varchar(255) not null,
		short   varchar(50)  not null,
		correlation_id varchar(100),
		is_deleted boolean default false
	);

	CREATE UNIQUE INDEX IF NOT EXISTS short_links_origin_uindex
    on public.short_links(origin);
	`

	if _, err := dbi.Exec(context.Background(), sql); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDatabaseExec, err)
	}

	return &PostgreSQLStorage{
		dbi:     dbi,
		l:       l,
		chBatch: ch,
	}, nil
}

// GetLinkDB get data from storage by short URL
func (s *PostgreSQLStorage) GetLinkDB(ctx context.Context, shortKey models.ShortURL) (models.Origin, error) {
	var origin models.Origin
	var gone bool

	query := "SELECT origin, is_deleted FROM public.short_links WHERE short=$1"
	err := s.dbi.QueryRow(ctx, query, string(shortKey)).Scan(&origin, &gone)

	if gone {
		return "", errs.ErrURLIsGone
	}

	if err != nil {
		return "", errs.ErrURLNotFound
	}

	return origin, nil
}

// LinksByUser return all user links
func (s *PostgreSQLStorage) LinksByUser(ctx context.Context, userID models.UniqUser) (models.ShortLinks, error) {
	query := "SELECT origin, short FROM public.short_links WHERE user_id=$1"

	origins := models.ShortLinks{}
	rows, err := s.dbi.Query(ctx, query, string(userID))
	if err != nil {
		return origins, err
	}

	for rows.Next() {
		var origin models.Origin
		var short models.ShortURL

		err = rows.Scan(&short, &origin)
		if err != nil {
			return origins, err
		}
		origins[short] = origin
	}

	return origins, nil
}

// SaveLinkDB save url in storage of short links
func (s *PostgreSQLStorage) SaveLinkDB(ctx context.Context, userID models.UniqUser, url models.Origin) (models.ShortURL, error) {

	shortKey, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	queryInsert := `
	INSERT INTO public.short_links (user_id, origin, short) 
	VALUES (@user_id, @origin, @short);
	`

	queryGet := `
	SELECT short FROM public.short_links where origin=$1
	`

	args := pgx.NamedArgs{
		"user_id": userID,
		"origin":  url,
		"short":   shortKey,
	}

	pgErr := &pgconn.PgError{}

	if _, err := s.dbi.Exec(ctx, queryInsert, args); err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				var short models.ShortURL
				_ = s.dbi.QueryRow(ctx, queryGet, string(url)).Scan(&short)

				return short, errs.ErrAlreadyHasShort
			}
		}
		return shortKey, nil
	}

	return shortKey, nil
}

// SaveBatch save multiply URL
func (s *PostgreSQLStorage) SaveBatch(ctx context.Context, userID models.UniqUser, urls []models.BatchReqURL) ([]models.BatchResURL, error) {
	type temp struct{ CorrID, Origin, Short string }

	var buffer []temp
	for _, v := range urls {
		shortKey, _ := utils.RandStringBytes()

		var t = temp{
			CorrID: v.CorrID,
			Origin: v.Origin,
			Short:  string(shortKey),
		}
		buffer = append(buffer, t)
	}

	var shorts []models.BatchResURL

	query := `
		INSERT INTO public.short_links (user_id, origin, short) 
		VALUES (@user_id, @origin, @short)
		ON CONFLICT (origin) DO NOTHING;
		`

	// Start transaction
	tx, err := s.dbi.Begin(ctx)
	defer tx.Rollback(ctx)
	if err != nil {
		return shorts, err
	}

	for _, v := range buffer {
		// Add record to transaction
		args := pgx.NamedArgs{
			"user_id":        userID,
			"origin":         v.Origin,
			"short":          v.Short,
			"correlation_id": v.CorrID,
		}

		if _, err = tx.Exec(ctx, query, args); err == nil {
			shorts = append(shorts, models.BatchResURL{
				Short:  v.Short,
				CorrID: v.CorrID,
			})
		} else {
			s.l.Info("Save bunch error", zap.Error(err))
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

// BunchUpdateAsDeleted delete mass URL by fanIN pattern
func (s *PostgreSQLStorage) BunchUpdateAsDeleted(chBatch chan models.BatchDelete) {
	for v := range chBatch {
		if len(v.URLs) == 0 {
			s.l.Info(errs.ErrCorrelation.Error())
		}

		query := `
		UPDATE public.short_links
		SET is_deleted=true
		WHERE user_id=$1 AND short=$2;
		`
		batch := &pgx.Batch{}

		for _, id := range v.URLs {
			batch.Queue(query, v.UserID, id)
		}

		results := s.dbi.SendBatch(context.Background(), batch)

		for _, id := range v.URLs {
			_, err := results.Exec()
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
					s.l.Info("update error on batch delete", zap.String("URL id", id))
					continue
				}
				s.l.Info("unable to delete row ", zap.Error(err))
			}
		}
		results.Close()
	}
}

// GetStats get statistics quantity urls and users
func (s *PostgreSQLStorage) GetStats(ctx context.Context, userID models.UniqUser) (models.GetStatsResURL, error) {
	stat := models.GetStatsResURL{}

	query := "SELECT count(distinct user_id) as users, count(origin) as urls FROM short_links;"
	err := s.dbi.QueryRow(ctx, query).Scan(&stat.URLs, &stat.Users)
	if err != nil {
		return stat, errs.ErrInternalSrv
	}

	return stat, nil
}
