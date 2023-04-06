package ramstorage

import (
	"errors"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/filestorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
)

var errNotFoundURL = errors.New("url not found in DB")

type RAMStorage struct {
	MU sync.Mutex
	DB map[user.UniqUser]storage.ShortLinks
}

func New() (*RAMStorage, error) {
	r := &RAMStorage{
		DB: make(map[user.UniqUser]storage.ShortLinks),
	}
	if err := r.Load(); err != nil {
		return r, err
	}
	return r, nil
}

func (r *RAMStorage) LinksByUser(userID user.UniqUser) (storage.ShortLinks, error) {
	shorts, ok := r.DB[userID]
	if !ok {
		return shorts, errNotFoundURL
	}

	return shorts, nil
}

func (r *RAMStorage) SaveLinkDB(userID user.UniqUser, url storage.ShortURL) (storage.URLKey, error) {
	r.MU.Lock()
	defer r.MU.Unlock()
	key, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	currentURL := storage.ShortLinks{}
	if urls, ok := r.DB[userID]; ok {
		currentURL = urls
	}

	currentURL[key] = url
	r.DB[userID] = currentURL
	r.DB["all"] = currentURL

	fs, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return key, nil
	}

	if err = filestorage.Write(fs, r.DB); err != nil {
		return "", err
	}

	return key, nil

}

func (r *RAMStorage) GetLinkDB(userID user.UniqUser, key storage.URLKey) (storage.ShortURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()
	shorts, ok := r.DB[userID]
	if !ok {
		return "", errNotFoundURL
	}

	url, ok := shorts[key]
	if !ok {
		return "", errNotFoundURL
	}

	return url, nil
}

// Load all links to map
func (r *RAMStorage) Load() error {
	fs, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return nil
	}
	if err := filestorage.Read(fs, &r.DB); err != nil {
		return err
	}
	return nil
}
