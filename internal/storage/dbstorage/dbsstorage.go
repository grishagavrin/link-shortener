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
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct{}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

func New() (*PostgreSQLStorage, error) {
	// Check if scheme exist
	sql := `
	CREATE TABLE IF NOT EXISTS public.short_links(
		id serial,
		user_id varchar(50),
		origin  varchar(255) not null,
		short   varchar(50)  not null,
		correlation_id varchar(100)
	);`

	if err := db.Insert(context.Background(), sql); err != nil {
		return &PostgreSQLStorage{}, err
	}

	return &PostgreSQLStorage{}, nil
}

func (s *PostgreSQLStorage) GetLinkDB(userID user.UniqUser, key storage.URLKey) (storage.ShortURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
	defer cancel()

	query := "select origin from public.short_links where short=$1"
	dbi, _ := db.Instance()

	var origin storage.ShortURL
	err := dbi.QueryRow(ctx, query, key).Scan(&origin)
	if err != nil {
		return "", ErrURLNotFound
	}

	return origin, nil
}

func (s *PostgreSQLStorage) LinksByUser(userID user.UniqUser) (storage.ShortLinks, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// не забываем освободить ресурс
	defer cancel()
	query := "select origin, short from public.short_links where user_id=$1"
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

	query := `
	INSERT INTO public.short_links (user_id, origin, short) 
	VALUES (@user_id, @origin, @short);
	`
	args := pgx.NamedArgs{
		"user_id": userID,
		"origin":  url,
		"short":   key,
	}

	err = db.Insert(ctx, query, args)

	if err != nil {
		fmt.Println("Unable to insert due to: ", err)
		return key, err
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
		var t = temp{
			ID:     v.ID,
			Origin: v.Origin,
			Short:  utils.RandomString(),
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
		VALUES (@user_id, @origin, @short, @correlation_id);`

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
