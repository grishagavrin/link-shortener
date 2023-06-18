package storage

import (
	"context"

	"github.com/grishagavrin/link-shortener/internal/user"
)

type ShortURL string
type Origin string
type ShortLinks map[Origin]ShortURL

type BatchURL struct {
	ID     string `json:"correlation_id"`
	Origin string `json:"original_url"`
}

type BatchShortURLs struct {
	Short string `json:"short_url"`
	ID    string `json:"correlation_id"`
}

type Repository interface {
	GetLinkDB(Origin) (ShortURL, error)
	SaveLinkDB(user.UniqUser, ShortURL) (Origin, error)
	SaveBatch(urls []BatchURL) ([]BatchShortURLs, error)
	LinksByUser(userID user.UniqUser) (ShortLinks, error)
	BunchUpdateAsDeleted(context.Context, []string, string) error
}
