package ramstorage

import (
	"context"
	"errors"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/filestorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"go.uber.org/zap"
)

var errNotFoundURL = errors.New("url not found in DB")

type RAMStorage struct {
	MU sync.Mutex
	DB map[user.UniqUser]storage.ShortLinks
	l  *zap.Logger
}

func New(l *zap.Logger) (*RAMStorage, error) {
	r := &RAMStorage{
		DB: make(map[user.UniqUser]storage.ShortLinks),
		l:  l,
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

func (r *RAMStorage) SaveLinkDB(userID user.UniqUser, url storage.ShortURL) (storage.Origin, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	r.l.Sugar().Infof("userID: ", string(userID))
	key, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	currentURLUser := storage.ShortLinks{}
	currentURLAll := storage.ShortLinks{}

	if urls, ok := r.DB[userID]; ok {
		currentURLUser = urls
	}

	currentURLUser[key] = url
	r.DB[userID] = currentURLUser

	if urls, ok := r.DB["all"]; ok {
		currentURLAll = urls
	}
	currentURLAll[key] = url
	r.DB["all"] = currentURLAll

	fs, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return key, nil
	}

	_ = filestorage.Write(fs, r.DB)
	return key, nil
}

func (r *RAMStorage) GetLinkDB(key storage.Origin) (storage.ShortURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()
	shorts, ok := r.DB["all"]

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
	// If file storage not exists
	if err != nil || fs == "" {
		return nil
	}

	if err := filestorage.Read(fs, &r.DB); err != nil {
		return err
	}
	return nil
}

// Batch save
func (r *RAMStorage) SaveBatch(urls []storage.BatchURL) ([]storage.BatchShortURLs, error) {
	var shorts []storage.BatchShortURLs
	return shorts, nil
}

func (r *RAMStorage) BunchUpdateAsDeleted(ctx context.Context, correlationIds []string, userID string) error {
	return nil
}
