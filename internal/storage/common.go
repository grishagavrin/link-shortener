package storage

import (
	"context"

	"github.com/grishagavrin/link-shortener/internal/user"
)

type ShortURL string
type Origin string
type ShortLinks map[Origin]ShortURL

type BatchReqURL struct {
	Corr_ID string `json:"correlation_id"`
	Origin  string `json:"original_url"`
}

type BatchResURL struct {
	Corr_ID string `json:"correlation_id"`
	Short   string `json:"short_url"`
}

type Repository interface {
	GetLinkDB(Origin) (ShortURL, error)
	SaveLinkDB(user.UniqUser, ShortURL) (Origin, error)
	SaveBatch(user.UniqUser, []BatchReqURL) ([]BatchResURL, error)
	LinksByUser(userID user.UniqUser) (ShortLinks, error)
	BunchUpdateAsDeleted(context.Context, []string, string) error
}
