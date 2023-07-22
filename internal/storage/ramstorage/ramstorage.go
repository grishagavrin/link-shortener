package ramstorage

import (
	"context"
	"fmt"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage/filestorage"
	"github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"go.uber.org/zap"
)

type RAMStorage struct {
	MU sync.Mutex
	DB map[user.UniqUser]iStorage.ShortLinks
	l  *zap.Logger
}

func New(l *zap.Logger) (*RAMStorage, error) {
	r := &RAMStorage{
		DB: make(map[user.UniqUser]iStorage.ShortLinks),
		l:  l,
	}

	if err := r.Load(); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrRamNotAvaliable, err)
	}
	return r, nil
}

func (r *RAMStorage) LinksByUser(_ context.Context, userID user.UniqUser) (iStorage.ShortLinks, error) {
	shorts, ok := r.DB[userID]
	if !ok {
		return shorts, errs.ErrNotFoundURL
	}

	return shorts, nil
}

func (r *RAMStorage) SaveLinkDB(_ context.Context, userID user.UniqUser, url iStorage.Origin) (iStorage.ShortURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	r.l.Sugar().Infof("userID: ", string(userID))
	shortKey, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	currentURLUser := iStorage.ShortLinks{}
	currentURLAll := iStorage.ShortLinks{}

	if urls, ok := r.DB[userID]; ok {
		currentURLUser = urls
	}

	currentURLUser[shortKey] = url
	r.DB[userID] = currentURLUser

	if urls, ok := r.DB["all"]; ok {
		currentURLAll = urls
	}
	currentURLAll[shortKey] = url
	r.DB["all"] = currentURLAll

	fs, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return shortKey, nil
	}

	_ = filestorage.Write(fs, r.DB)
	return shortKey, nil
}

func (r *RAMStorage) GetLinkDB(_ context.Context, key iStorage.ShortURL) (iStorage.Origin, error) {
	r.MU.Lock()
	defer r.MU.Unlock()
	shorts, ok := r.DB["all"]

	if !ok {
		return "", errs.ErrNotFoundURL
	}

	url, ok := shorts[key]
	if !ok {
		return "", errs.ErrNotFoundURL
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
func (r *RAMStorage) SaveBatch(_ context.Context, userID user.UniqUser, urls []iStorage.BatchReqURL) ([]iStorage.BatchResURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	var shortsRes []iStorage.BatchResURL

	currentURLUser := iStorage.ShortLinks{}
	currentURLAll := iStorage.ShortLinks{}

	r.l.Sugar().Infof("userID: ", string(userID))

	for _, url := range urls {
		shortKey, _ := utils.RandStringBytes()

		resItem := iStorage.BatchResURL{
			CorrID: url.CorrID,
			Short:  string(shortKey),
		}

		if urls, ok := r.DB[userID]; ok {
			currentURLUser = urls
		}

		currentURLUser[shortKey] = iStorage.Origin(url.Origin)
		r.DB[userID] = currentURLUser

		if urls, ok := r.DB["all"]; ok {
			currentURLAll = urls
		}

		currentURLAll[shortKey] = iStorage.Origin(url.Origin)
		r.DB["all"] = currentURLAll
		shortsRes = append(shortsRes, resItem)
	}

	fs, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return shortsRes, nil
	}

	_ = filestorage.Write(fs, r.DB)
	return shortsRes, nil
}

func (r *RAMStorage) BunchUpdateAsDeleted(ctx context.Context, correlationIds []string, userID string) error {
	r.MU.Lock()
	defer r.MU.Unlock()

	if len(correlationIds) == 0 {
		return errs.ErrCorrelation
	}

	for _, v := range correlationIds {
		delete(r.DB[user.UniqUser(userID)], iStorage.ShortURL(v))
		delete(r.DB["all"], iStorage.ShortURL(v))
	}

	return nil
}
