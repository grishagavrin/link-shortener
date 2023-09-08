// Package db consist function for work with database connection objects
package db

import (
	"context"
	"errors"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var instance *pgxpool.Pool

// Instance connection instance
func Instance(l *zap.Logger) (*pgxpool.Pool, error) {
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
			return instance, errs.ErrDatabaseNotAvaliable
		}

		inst, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			return instance, err
		}

		instance = inst
		l.Info("Connecting to DB")
	}

	return instance, nil
}
