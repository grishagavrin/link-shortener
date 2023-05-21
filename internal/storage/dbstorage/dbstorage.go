package dbstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"github.com/grishagavrin/link-shortener/internal/utils/db"
	"github.com/jackc/pgx/v5"
)

// PostgreSQLStorage storage
type PostgreSQLStorage struct{}

// ErrURLNotFound error by package level
var ErrURLNotFound = errors.New("url not found")

func New() (*PostgreSQLStorage, error) {
	// Check if scheme exist
	sql := `
	CREATE TABLE IF NOT EXISTS storage.short_links(
		id serial,
		user_id varchar(50),
		origin  varchar(255) not null,
		short   varchar(50)  not null
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

	query := "select origin from storage.short_links where short=$1"
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
	query := "select origin, short from storage.short_links where user_id=$1"
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
	INSERT INTO storage.short_links (user_id, origin, short) 
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
