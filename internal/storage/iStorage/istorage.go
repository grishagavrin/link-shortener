// Package istorage implements Repository pattern
package istorage

import (
	"context"
)

// UniqUser unique user type
type UniqUser string

// ShortURL short url type
type ShortURL string

// Origin original url type
type Origin string

// ShortLinks map of shorturl/origin types
type ShortLinks map[ShortURL]Origin

// OriginRAM for bool delete in origin
type OriginRAM struct {
	Origin    Origin
	IsDeleted bool
}

// ShortLinksRAM RAM storage
type ShortLinksRAM map[ShortURL]OriginRAM

// BatchDelete response struct
type BatchDelete struct {
	UserID string
	URLs   []string
}

// BatchReqURL request
type BatchReqURL struct {
	CorrID string `json:"correlation_id" example:"1237978947"`
	Origin string `json:"original_url" example:"http://yandex.ru"`
}

// BatchResURL response
type BatchResURL struct {
	CorrID string `json:"correlation_id"`
	Short  string `json:"short_url"`
}

// GetStatsReqURL request
type GetStatsReqURL struct {
	URLs  int `json:"urls" example:"12"`
	Users int `json:"users" example:"5"`
}

// Repository interface for working with global repository
type Repository interface {
	GetLinkDB(context.Context, ShortURL) (Origin, error)
	SaveLinkDB(context.Context, UniqUser, Origin) (ShortURL, error)
	LinksByUser(context.Context, UniqUser) (ShortLinks, error)
	SaveBatch(context.Context, UniqUser, []BatchReqURL) ([]BatchResURL, error)
	BunchUpdateAsDeleted(chan BatchDelete)
	GetStats(context.Context, UniqUser) (GetStatsReqURL, error)
}
