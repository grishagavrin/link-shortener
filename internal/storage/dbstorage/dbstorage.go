package dbstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"github.com/grishagavrin/link-shortener/internal/utils/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct{}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")
var ErrAlreadyHasShort = errors.New("already has short")
var ErrURLIsGone = errors.New("url is gone")

func New() (*PostgreSQLStorage, error) {
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

	if err := db.Insert(context.Background(), sql); err != nil {
		return &PostgreSQLStorage{}, err
	}

	return &PostgreSQLStorage{}, nil
}

func (s *PostgreSQLStorage) GetLinkDB(key storage.URLKey) (storage.ShortURL, error) {
	dbi, _ := db.Instance()
	var origin storage.ShortURL
	var gone bool

	query := "select origin, is_deleted from public.short_links where short=$1"

	err := dbi.QueryRow(context.Background(), query, string(key)).Scan(&origin, &gone)

	if err != nil {
		return "", ErrURLNotFound
	}

	if gone {
		return "", ErrURLIsGone
	}

	return origin, nil
}

func (s *PostgreSQLStorage) LinksByUser(userID user.UniqUser) (storage.ShortLinks, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
	defer cancel()
	query := "SELECT origin, short FROM public.short_links WHERE user_id=$1"
	dbi, _ := db.Instance()

	origins := storage.ShortLinks{}
	rows, err := dbi.Query(ctx, query, string(userID))
	if err != nil {
		return origins, err
	}

	for rows.Next() {
		var origin storage.URLKey
		var short storage.ShortURL

		err = rows.Scan(&short, &origin)
		if err != nil {
			return origins, err
		}
		origins[origin] = short
	}

	return origins, nil
}

// Save url in storage of short links
func (s *PostgreSQLStorage) SaveLinkDB(userID user.UniqUser, url storage.ShortURL) (storage.URLKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
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

	dbi, _ := db.Instance()
	pgErr := &pgconn.PgError{}

	if _, err := dbi.Exec(ctx, queryInsert, args); err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				var short storage.URLKey
				_ = dbi.QueryRow(ctx, queryGet, string(userID), url).Scan(&short)
				return short, ErrAlreadyHasShort
			}
		}
		return key, nil
	}

	return key, nil
}

// Save url in storage of short links
func (s *PostgreSQLStorage) SaveBatch(urls []storage.BatchURL) ([]storage.BatchShortURLs, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
	defer cancel()
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

	dbi, _ := db.Instance()
	var shorts []storage.BatchShortURLs
	// Delete old records for tests
	_, _ = dbi.Exec(ctx, "truncate table public.short_links;")

	// sqlBunchNewRecord for new record in db
	query := `
		INSERT INTO public.short_links (user_id, origin, short, correlation_id) 
		VALUES (@user_id, @origin, @short, @correlation_id)
		ON CONFLICT (user_id, origin) DO NOTHING;
		`

	// Start transaction
	tx, err := dbi.Begin(ctx)
	if err != nil {
		return shorts, err
	}
	defer tx.Rollback(ctx)

	for _, v := range buffer {
		// Add record to transaction
		args := pgx.NamedArgs{
			"user_id":        "all",
			"origin":         v.Origin,
			"short":          v.Short,
			"correlation_id": v.ID,
		}

		if _, err = tx.Exec(ctx, query, args); err == nil {
			shorts = append(shorts, storage.BatchShortURLs{
				Short: v.Short,
				ID:    v.ID,
			})
		} else {
			logger.Info("Save bunch error", zap.Error(err))
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

func BunchUpdateAsDeleted(ctx context.Context, correlationIds string, userID string) (string, error) {
	dbi, _ := db.Instance()
	if len(correlationIds) == 0 {
		return "correlationIds is null", nil
	}

	tx, err := dbi.Begin(ctx)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return "ERROR: tx begin error", err
	}
	defer tx.Rollback(ctx)

	query := `
	UPDATE public.short_links 
	SET is_deleted=true 
	WHERE user_id=$1
	AND (correlation_id =($2) OR short=($3))
	RETURNING short
	`

	if _, err = tx.Exec(ctx, query, userID, correlationIds, correlationIds); err != nil {
		fmt.Println(err)
		return "tx exec error", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "tx commit error", err
	}

	return correlationIds, nil
}
