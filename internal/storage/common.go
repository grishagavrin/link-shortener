package storage

type ShortURL string
type URLKey string

type Repository interface {
	GetLinkDB(URLKey) (ShortURL, error)
	SaveLinkDB(ShortURL) (URLKey, error)
}
