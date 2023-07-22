package dbstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct {
	dbi *pgxpool.Pool
	l   *zap.Logger
}

func New(dbi *pgxpool.Pool, l *zap.Logger) (*PostgreSQLStorage, error) {
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

	CREATE UNIQUE INDEX IF NOT EXISTS short_links_user_id_origin_uindex
    on public.short_links (user_id, origin);
	`

	if _, err := dbi.Exec(context.Background(), sql); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrDatabaseExec, err)
	}

	return &PostgreSQLStorage{
		dbi: dbi,
		l:   l,
	}, nil
}

func (s *PostgreSQLStorage) GetLinkDB(ctx context.Context, shortKey iStorage.ShortURL) (iStorage.Origin, error) {
	var origin iStorage.Origin
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

func (s *PostgreSQLStorage) LinksByUser(ctx context.Context, userID user.UniqUser) (iStorage.ShortLinks, error) {
	query := "SELECT origin, short FROM public.short_links WHERE user_id=$1"

	origins := iStorage.ShortLinks{}
	rows, err := s.dbi.Query(ctx, query, string(userID))
	if err != nil {
		return origins, err
	}

	for rows.Next() {
		var origin iStorage.Origin
		var short iStorage.ShortURL

		err = rows.Scan(&short, &origin)
		if err != nil {
			return origins, err
		}
		origins[short] = origin
	}

	return origins, nil
}

// Save url
func (s *PostgreSQLStorage) SaveLinkDB(ctx context.Context, userID user.UniqUser, url iStorage.Origin) (iStorage.ShortURL, error) {

	shortKey, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	queryInsert := `
	INSERT INTO public.short_links (user_id, origin, short) 
	VALUES (@user_id, @origin, @short);
	`

	queryGet := `
	SELECT short FROM public.short_links where user_id=$1 and origin=$2
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
				var short iStorage.ShortURL
				_ = s.dbi.QueryRow(ctx, queryGet, string(userID), url).Scan(&short)
				return short, errs.ErrAlreadyHasShort
			}
		}
		return shortKey, nil
	}

	return shortKey, nil
}

// Save url batch
func (s *PostgreSQLStorage) SaveBatch(ctx context.Context, userID user.UniqUser, urls []iStorage.BatchReqURL) ([]iStorage.BatchResURL, error) {
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

	var shorts []iStorage.BatchResURL

	// Delete old records
	_, _ = s.dbi.Exec(ctx, "TRUNCATE TABLE public.short_links;")

	query := `
		INSERT INTO public.short_links (user_id, origin, short, correlation_id) 
		VALUES (@user_id, @origin, @short, @correlation_id)
		ON CONFLICT (user_id, origin) DO NOTHING;
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
			shorts = append(shorts, iStorage.BatchResURL{
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

func (s *PostgreSQLStorage) BunchUpdateAsDeleted(ctx context.Context, shortIds []string, userID string) error {
	if len(shortIds) == 0 {
		return errs.ErrCorrelation
	}

	query := `
	UPDATE public.short_links 
	SET is_deleted=true 
	WHERE user_id=$1 AND short=$2;
	`
	batch := &pgx.Batch{}

	for _, id := range shortIds {
		batch.Queue(query, userID, id)
	}

	results := s.dbi.SendBatch(ctx, batch)
	defer results.Close()

	for _, id := range shortIds {
		_, err := results.Exec()
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				s.l.Sugar().Infof("update error on %s", id)
				continue
			}
			return fmt.Errorf("unable to insert row: %w", err)
		}
	}
	return results.Close()
}
