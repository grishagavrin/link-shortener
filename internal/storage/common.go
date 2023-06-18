package storage

import (
	"github.com/grishagavrin/link-shortener/internal/user"
)

type ShortURL string
type URLKey string
type ShortLinks map[URLKey]ShortURL
type BatchURL struct {
	ID     string `json:"correlation_id"`
	Origin string `json:"original_url"`
}
type BatchShortURLs struct {
	Short string `json:"short_url"`
	ID    string `json:"correlation_id"`
}

type Repository interface {
	GetLinkDB(user.UniqUser, URLKey) (ShortURL, error)
	SaveLinkDB(user.UniqUser, ShortURL) (URLKey, error)
	SaveBatch(urls []BatchURL) ([]BatchShortURLs, error)
	LinksByUser(userID user.UniqUser) (ShortLinks, error)
}
