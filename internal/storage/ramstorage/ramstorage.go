package ramstorage

import (
	"errors"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/utils"
)

var errNotFoundURL = errors.New("url not found in DB")

type RAMStorage struct {
	MU sync.Mutex
	DB map[storage.URLKey]storage.ShortURL
}

func New() *RAMStorage {
	return &RAMStorage{
		DB: make(map[storage.URLKey]storage.ShortURL),
	}
}

func (r *RAMStorage) SaveLinkDB(url storage.ShortURL) (storage.URLKey, error) {
	r.MU.Lock()
	defer r.MU.Unlock()
	key := utils.RandStringBytes(config.LENHASH)

	if _, ok := r.DB[key]; !ok {
		r.DB[key] = url
	}

	return key, nil
}

func (r *RAMStorage) GetLinkDB(key storage.URLKey) (storage.ShortURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()
	v, ok := r.DB[key]
	if !ok {
		return v, errNotFoundURL
	}
	return v, nil
}
