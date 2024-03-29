// Package filestorage contains methods for file storage
package filestorage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage/filewrapper"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"go.uber.org/zap"
)

// RAMStorage for file storage
type RAMStorage struct {
	MU      sync.Mutex
	DB      map[models.UniqUser]models.ShortLinksRAM
	l       *zap.Logger
	chBatch chan models.BatchDelete
}

// New instance new storage wit not null fields
func New(l *zap.Logger, ch chan models.BatchDelete) (*RAMStorage, error) {
	r := &RAMStorage{
		DB:      make(map[models.UniqUser]models.ShortLinksRAM),
		l:       l,
		chBatch: ch,
	}

	if err := r.Load(); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrRAMNotAvaliable, err)
	}
	return r, nil
}

// Load all links to map
func (r *RAMStorage) Load() error {
	// fs, err := config.Instance().GetCfgValue(config.FileStoragePath)
	// Config instance
	cfg, _ := config.Instance()
	// Config value
	fs, err := cfg.GetCfgValue(config.FileStoragePath)

	// If file storage not exists
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		return nil
	}

	if err := filewrapper.Read(fs, &r.DB); err != nil {
		return err
	}
	return nil
}

// LinksByUser return all user links
func (r *RAMStorage) LinksByUser(_ context.Context, userID models.UniqUser) (models.ShortLinks, error) {
	shorts := models.ShortLinks{}
	shortsRAM, ok := r.DB[userID]
	if !ok {
		return nil, errs.ErrNotFoundURL
	}

	for k, v := range shortsRAM {
		shorts[k] = v.Origin
	}

	return shorts, nil
}

// SaveLinkDB save url in storage of short links
func (r *RAMStorage) SaveLinkDB(_ context.Context, userID models.UniqUser, url models.Origin) (models.ShortURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	shortKey, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	currentURLUserRes := models.ShortLinks{}
	currentURLUserRAM := models.ShortLinksRAM{}

	if urls, ok := r.DB[userID]; ok {
		for k, v := range urls {
			if v.Origin == url {
				return k, errs.ErrAlreadyHasShort
			}
		}
		currentURLUserRAM = urls
	}

	currentURLUserRAM[shortKey] = models.OriginRAM{
		Origin:    url,
		IsDeleted: false,
	}
	currentURLUserRes[shortKey] = url

	r.DB[userID] = currentURLUserRAM

	currentURLAllRes := models.ShortLinks{}
	currentURLAllRAM := models.ShortLinksRAM{}

	if urls, ok := r.DB["all"]; ok {
		for k, v := range urls {
			if v.Origin == url {
				return k, errs.ErrAlreadyHasShort
			}
		}
		currentURLAllRAM = urls
	}

	currentURLAllRAM[shortKey] = models.OriginRAM{
		Origin:    url,
		IsDeleted: false,
	}
	currentURLAllRes[shortKey] = url

	r.DB["all"] = currentURLAllRAM

	// Config instance
	cfg, _ := config.Instance()
	// Config value
	fs, err := cfg.GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return "", errs.ErrUnknownEnvOrFlag
	}

	_ = filewrapper.Write(fs, r.DB)

	return shortKey, nil
}

// GetLinkDB get data from storage by short URL
func (r *RAMStorage) GetLinkDB(_ context.Context, key models.ShortURL) (models.Origin, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	allShorts, ok := r.DB["all"]
	if !ok {
		return "", errs.ErrNotFoundURL
	}

	originRAM, ok := allShorts[key]
	if ok && originRAM.IsDeleted {
		return "", errs.ErrURLIsGone
	} else if !ok {
		return "", errs.ErrURLNotFound
	}

	return originRAM.Origin, nil
}

// SaveBatch save multiply URL
func (r *RAMStorage) SaveBatch(_ context.Context, userID models.UniqUser, urls []models.BatchReqURL) ([]models.BatchResURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	var shortsRes []models.BatchResURL

	currentURLUserRAM := models.ShortLinksRAM{}
	currentURLAllRAM := models.ShortLinksRAM{}

	if r.DB["all"] != nil && r.DB[userID] != nil {
		currentURLAllRAM = r.DB["all"]
		currentURLUserRAM = r.DB[userID]
	}

	for _, url := range urls {
		shortKey, _ := utils.RandStringBytes()

		currentURLUserRAM[shortKey] = models.OriginRAM{
			Origin:    models.Origin(url.Origin),
			IsDeleted: false,
		}

		r.DB[userID] = currentURLUserRAM

		currentURLAllRAM[shortKey] = models.OriginRAM{
			Origin:    models.Origin(url.Origin),
			IsDeleted: false,
		}

		r.DB["all"] = currentURLAllRAM

		resItem := models.BatchResURL{
			CorrID: url.CorrID,
			Short:  string(shortKey),
		}

		shortsRes = append(shortsRes, resItem)
	}

	// Config instance
	cfg, _ := config.Instance()
	// Config value
	fs, err := cfg.GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return nil, errs.ErrUnknownEnvOrFlag
	}

	_ = filewrapper.Write(fs, r.DB)

	return shortsRes, nil
}

// BunchUpdateAsDeleted delete mass URL by fanIN pattern
func (r *RAMStorage) BunchUpdateAsDeleted(chBatch chan models.BatchDelete) {
	for v := range chBatch {
		r.MU.Lock()

		// Config instance
		cfg, _ := config.Instance()
		// Config value
		fs, err := cfg.GetCfgValue(config.FileStoragePath)
		if err != nil || fs == "" {
			r.l.Info(errs.ErrUnknownEnvOrFlag.Error())
		}

		if len(v.URLs) == 0 {
			r.l.Info(errs.ErrCorrelation.Error())
		}

		shortUser := r.DB[models.UniqUser(v.UserID)]
		shortAll := r.DB["all"]

		for _, v := range v.URLs {
			if su, ok := shortUser[models.ShortURL(v)]; ok {
				su.IsDeleted = true
				shortUser[models.ShortURL(v)] = su
			}

			if sa, ok := shortAll[models.ShortURL(v)]; ok {
				sa.IsDeleted = true
				shortAll[models.ShortURL(v)] = sa
			}
		}
		_ = filewrapper.Write(fs, r.DB)
		r.MU.Unlock()
	}
}

// GetStats get statistics quantity urls and users
func (r *RAMStorage) GetStats(_ context.Context, userID models.UniqUser) (models.GetStatsResURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	return models.GetStatsResURL{
		Users: len(r.DB) - 1,    // because r.DB["all"] must be deleted
		URLs:  len(r.DB["all"]), // r.DB["all"] contains all hashed links
	}, nil
}
