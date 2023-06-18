package db

import (
	"context"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var instance *pgxpool.Pool

func Instance(l *zap.Logger) (*pgxpool.Pool, error) {
	if instance == nil {
		dsn, _ := config.Instance().GetCfgValue(config.DatabaseDSN)
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
