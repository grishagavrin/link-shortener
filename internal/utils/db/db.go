package db

import (
	"context"
	"errors"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDatabaseNotAvaliable = errors.New("db not avaliable")

var instance *pgxpool.Pool

func Instance() (*pgxpool.Pool, error) {
	if instance == nil {
		dsn, _ := config.Instance().GetCfgValue(config.DatabaseDSN)
		if dsn == "" {
			return instance, ErrDatabaseNotAvaliable
		}

		inst, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			return instance, err
		}

		instance = inst
		logger.Info("Connecting to DB")
	}
	return instance, nil
}
