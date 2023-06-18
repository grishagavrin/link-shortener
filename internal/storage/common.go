package storage

import (
	"context"

	"github.com/grishagavrin/link-shortener/internal/user"
)

type ShortURL string
type Origin string
type ShortLinks map[ShortURL]Origin

type BatchReqURL struct {
	CorrID string `json:"correlation_id"`
	Origin string `json:"original_url"`
}

type BatchResURL struct {
	CorrID string `json:"correlation_id"`
	Short  string `json:"short_url"`
}

type Repository interface {
	GetLinkDB(ShortURL) (Origin, error)
	SaveLinkDB(user.UniqUser, Origin) (ShortURL, error)
	SaveBatch(user.UniqUser, []BatchReqURL) ([]BatchResURL, error)
	LinksByUser(userID user.UniqUser) (ShortLinks, error)
	BunchUpdateAsDeleted(context.Context, []string, string) error
}
