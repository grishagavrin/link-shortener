package dbstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"github.com/grishagavrin/link-shortener/internal/utils/db"
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

func New(l *zap.Logger) (*PostgreSQLStorage, error) {
	// Init DB
	dbi, _ := db.Instance(l)
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
		return &PostgreSQLStorage{}, err
	}

	return &PostgreSQLStorage{
		dbi: dbi,
		l:   l,
	}, nil
}

func (s *PostgreSQLStorage) GetLinkDB(key storage.Origin) (storage.ShortURL, error) {
	var origin storage.ShortURL
	var gone bool

	query := "SELECT origin, is_deleted FROM public.short_links WHERE short=$1"
	err := s.dbi.QueryRow(context.Background(), query, string(key)).Scan(&origin, &gone)

	if gone {
		return "", errs.ErrURLIsGone
	}

	if err != nil {
		return "", errs.ErrURLNotFound
	}

	return origin, nil
}

func (s *PostgreSQLStorage) LinksByUser(userID user.UniqUser) (storage.ShortLinks, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
	defer cancel()
	query := "SELECT origin, short FROM public.short_links WHERE user_id=$1"

	origins := storage.ShortLinks{}
	rows, err := s.dbi.Query(ctx, query, string(userID))
	if err != nil {
		return origins, err
	}

	for rows.Next() {
		var origin storage.Origin
		var short storage.ShortURL

		err = rows.Scan(&short, &origin)
		if err != nil {
			return origins, err
		}
		origins[origin] = short
	}

	return origins, nil
}

// Save url
func (s *PostgreSQLStorage) SaveLinkDB(userID user.UniqUser, url storage.ShortURL) (storage.Origin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key, err := utils.RandStringBytes()
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
		"short":   key,
	}

	pgErr := &pgconn.PgError{}

	if _, err := s.dbi.Exec(ctx, queryInsert, args); err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				var short storage.Origin
				_ = s.dbi.QueryRow(ctx, queryGet, string(userID), url).Scan(&short)
				return short, errs.ErrAlreadyHasShort
			}
		}
		return key, nil
	}

	return key, nil
}

// Save url batch
func (s *PostgreSQLStorage) SaveBatch(userID user.UniqUser, urls []storage.BatchURL) ([]storage.BatchShortURLs, error) {
	type temp struct {
		ID,
		Origin,
		Short string
	}

	var buffer []temp
	for _, v := range urls {
		key, _ := utils.RandStringBytes()

		var t = temp{
			ID:     v.ID,
			Origin: v.Origin,
			Short:  string(key),
		}
		buffer = append(buffer, t)
	}

	var shorts []storage.BatchShortURLs
	// Delete old records for tests
	// _, _ = s.dbi.Exec(context.Background(), "TRUNCATE TABLE public.short_links;")

	query := `
		INSERT INTO public.short_links (user_id, origin, short, correlation_id) 
		VALUES (@user_id, @origin, @short, @correlation_id)
		ON CONFLICT (user_id, origin) DO NOTHING;
		`

	// Start transaction
	tx, err := s.dbi.Begin(context.Background())
	if err != nil {
		return shorts, err
	}

	defer tx.Rollback(context.Background())

	for _, v := range buffer {
		// Add record to transaction
		args := pgx.NamedArgs{
			"user_id":        userID,
			"origin":         v.Origin,
			"short":          v.Short,
			"correlation_id": v.ID,
		}

		if _, err = tx.Exec(context.Background(), query, args); err == nil {
			shorts = append(shorts, storage.BatchShortURLs{
				Short: v.Short,
				ID:    v.ID,
			})
		} else {
			s.l.Info("Save bunch error", zap.Error(err))
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

func (s *PostgreSQLStorage) BunchUpdateAsDeleted(ctx context.Context, correlationIds []string, userID string) error {
	if len(correlationIds) == 0 {
		return errs.ErrCorrelation
	}

	query := `
	UPDATE public.short_links 
	SET is_deleted=true 
	WHERE user_id=$1 AND short=$2;
	`
	batch := &pgx.Batch{}

	for _, id := range correlationIds {
		batch.Queue(query, userID, id)
	}

	results := s.dbi.SendBatch(ctx, batch)
	defer results.Close()

	for _, id := range correlationIds {
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
