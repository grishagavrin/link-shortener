package storage

import (
	"github.com/grishagavrin/link-shortener/internal/user"
)

type ShortURL string
type URLKey string

type ShortLinks map[URLKey]ShortURL

type Repository interface {
	GetLinkDB(user.UniqUser, URLKey) (ShortURL, error)
	SaveLinkDB(user.UniqUser, ShortURL) (URLKey, error)
	LinksByUser(userID user.UniqUser) (ShortLinks, error)
}
