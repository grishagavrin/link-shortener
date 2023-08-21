package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage/dbstorage"
	istorage "github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"github.com/grishagavrin/link-shortener/internal/storage/ramstorage"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var instance *pgxpool.Pool

func SQLDBConnection(l *zap.Logger) (*pgxpool.Pool, error) {
	if instance == nil {
		// Config instance
		cfg, err := config.Instance()
		if errors.Is(err, errs.ErrENVLoading) {
			return nil, errs.ErrDatabaseNotAvaliable
		}

		//Config value
		dsn, err := cfg.GetCfgValue(config.DatabaseDSN)
		if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
			return nil, errs.ErrDatabaseNotAvaliable
		}

		if dsn == "" {
			return nil, errs.ErrDatabaseNotAvaliable
		}

		inst, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", errs.ErrDatabaseNotAvaliable, err)
		}

		instance = inst
	}

	return instance, nil
}

func Instance(l *zap.Logger) (istorage.Repository, *pgxpool.Pool, error) {
	dbi, err := SQLDBConnection(l)
	if err == nil {
		stor, err := dbstorage.New(dbi, l)
		if errors.Is(err, errs.ErrDatabaseNotAvaliable) || errors.Is(err, errs.ErrDatabaseExec) {
			return nil, dbi, err
		}

		l.Info("Connected to DB")
		return stor, dbi, nil
	} else {
		stor, err := ramstorage.New(l)
		if errors.Is(err, errs.ErrRAMNotAvaliable) {
			return nil, nil, err
		}

		l.Info("Set RAM handler")
		return stor, nil, nil
	}
}
