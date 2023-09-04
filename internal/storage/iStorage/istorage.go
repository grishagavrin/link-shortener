// Package istorage implements Repository pattern
package istorage

import (
	"context"
)

type UniqUser string
type ShortURL string
type Origin string
type ShortLinks map[ShortURL]Origin

type OriginRAM struct {
	Origin    Origin
	IsDeleted bool
}
type ShortLinksRAM map[ShortURL]OriginRAM

type BatchDelete struct {
	UserID string
	URLs   []string
}

type BatchReqURL struct {
	CorrID string `json:"correlation_id"`
	Origin string `json:"original_url"`
}

type BatchResURL struct {
	CorrID string `json:"correlation_id"`
	Short  string `json:"short_url"`
}

type Repository interface {
	GetLinkDB(context.Context, ShortURL) (Origin, error)
	SaveLinkDB(context.Context, UniqUser, Origin) (ShortURL, error)
	LinksByUser(context.Context, UniqUser) (ShortLinks, error)
	SaveBatch(context.Context, UniqUser, []BatchReqURL) ([]BatchResURL, error)
	BunchUpdateAsDeleted(chan BatchDelete)
}
