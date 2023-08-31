package ramstorage

import (
	"context"
	"fmt"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage/filewrapper"
	istorage "github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"go.uber.org/zap"
)

type RAMStorage struct {
	MU      sync.Mutex
	DB      map[user.UniqUser]istorage.ShortLinksRAM
	l       *zap.Logger
	chBatch chan istorage.BatchDelete
}

func New(l *zap.Logger, ch chan istorage.BatchDelete) (*RAMStorage, error) {
	r := &RAMStorage{
		DB:      make(map[user.UniqUser]istorage.ShortLinksRAM),
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
	//Config value
	fs, err := cfg.GetCfgValue(config.FileStoragePath)

	// If file storage not exists
	if err != nil || fs == "" {
		return nil
	}

	if err := filewrapper.Read(fs, &r.DB); err != nil {
		return err
	}
	return nil
}

func (r *RAMStorage) LinksByUser(_ context.Context, userID user.UniqUser) (istorage.ShortLinks, error) {
	shorts := istorage.ShortLinks{}
	shortsRAM, ok := r.DB[userID]
	if !ok {
		return nil, errs.ErrNotFoundURL
	}

	for k, v := range shortsRAM {
		shorts[k] = v.Origin
	}

	return shorts, nil
}

func (r *RAMStorage) SaveLinkDB(_ context.Context, userID user.UniqUser, url istorage.Origin) (istorage.ShortURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	shortKey, err := utils.RandStringBytes()
	if err != nil {
		return "", err
	}

	currentURLUserRes := istorage.ShortLinks{}
	currentURLUserRAM := istorage.ShortLinksRAM{}

	if urls, ok := r.DB[userID]; ok {
		for k, v := range urls {
			if v.Origin == url {
				return k, errs.ErrAlreadyHasShort
			}
		}
		currentURLUserRAM = urls
	}

	currentURLUserRAM[shortKey] = istorage.OriginRAM{
		Origin:    url,
		IsDeleted: false,
	}
	currentURLUserRes[shortKey] = url

	r.DB[userID] = currentURLUserRAM

	currentURLAllRes := istorage.ShortLinks{}
	currentURLAllRAM := istorage.ShortLinksRAM{}

	if urls, ok := r.DB["all"]; ok {
		for k, v := range urls {
			if v.Origin == url {
				return k, errs.ErrAlreadyHasShort
			}
		}
		currentURLAllRAM = urls
	}

	currentURLAllRAM[shortKey] = istorage.OriginRAM{
		Origin:    url,
		IsDeleted: false,
	}
	currentURLAllRes[shortKey] = url

	r.DB["all"] = currentURLAllRAM

	// Config instance
	cfg, _ := config.Instance()
	//Config value
	fs, err := cfg.GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return "", errs.ErrUnknownEnvOrFlag
	}

	_ = filewrapper.Write(fs, r.DB)

	return shortKey, nil
}

func (r *RAMStorage) GetLinkDB(_ context.Context, key istorage.ShortURL) (istorage.Origin, error) {
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

// Batch save
func (r *RAMStorage) SaveBatch(_ context.Context, userID user.UniqUser, urls []istorage.BatchReqURL) ([]istorage.BatchResURL, error) {
	r.MU.Lock()
	defer r.MU.Unlock()

	var shortsRes []istorage.BatchResURL

	currentURLUserRAM := istorage.ShortLinksRAM{}
	currentURLAllRAM := istorage.ShortLinksRAM{}

	if r.DB["all"] != nil && r.DB[userID] != nil {
		currentURLAllRAM = r.DB["all"]
		currentURLUserRAM = r.DB[userID]
	}

	for _, url := range urls {
		shortKey, _ := utils.RandStringBytes()

		currentURLUserRAM[shortKey] = istorage.OriginRAM{
			Origin:    istorage.Origin(url.Origin),
			IsDeleted: false,
		}

		r.DB[userID] = currentURLUserRAM

		currentURLAllRAM[shortKey] = istorage.OriginRAM{
			Origin:    istorage.Origin(url.Origin),
			IsDeleted: false,
		}

		r.DB["all"] = currentURLAllRAM

		resItem := istorage.BatchResURL{
			CorrID: url.CorrID,
			Short:  string(shortKey),
		}

		shortsRes = append(shortsRes, resItem)
	}

	// Config instance
	cfg, _ := config.Instance()
	//Config value
	fs, err := cfg.GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return nil, errs.ErrUnknownEnvOrFlag
	}

	_ = filewrapper.Write(fs, r.DB)

	return shortsRes, nil
}

func (r *RAMStorage) BunchUpdateAsDeleted(chBatch chan istorage.BatchDelete) {
	for v := range chBatch {
		r.MU.Lock()

		// Config instance
		cfg, _ := config.Instance()
		//Config value
		fs, err := cfg.GetCfgValue(config.FileStoragePath)
		if err != nil || fs == "" {
			r.l.Info(errs.ErrUnknownEnvOrFlag.Error())
		}

		if len(v.URLs) == 0 {
			r.l.Info(errs.ErrCorrelation.Error())
		}

		shortUser := r.DB[user.UniqUser(v.UserID)]
		shortAll := r.DB["all"]

		for _, v := range v.URLs {
			if su, ok := shortUser[istorage.ShortURL(v)]; ok {
				su.IsDeleted = true
				shortUser[istorage.ShortURL(v)] = su
			}

			if sa, ok := shortAll[istorage.ShortURL(v)]; ok {
				sa.IsDeleted = true
				shortAll[istorage.ShortURL(v)] = sa
			}
		}
		_ = filewrapper.Write(fs, r.DB)
		r.MU.Unlock()
	}
}
