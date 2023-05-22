package db

import (
	"context"
	"errors"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/jackc/pgx/v5"
)

var ErrDatabaseNotAvaliable = errors.New("db not avaliable")

var instance *pgx.Conn

func Instance() (*pgx.Conn, error) {
	if instance == nil {
		dsn, _ := config.Instance().GetCfgValue(config.DatabaseDSN)
		if dsn == "" {
			return instance, ErrDatabaseNotAvaliable
		}

		inst, err := pgx.Connect(context.Background(), dsn)
		if err != nil {
			return instance, err
		}
		instance = inst
		logger.Info("Connecting to DB")
	}

	return instance, nil
}

// Insert execute query to active connect
func Insert(ctx context.Context, query string, args ...interface{}) error {
	c, err := Instance()
	if err == nil {
		if _, err := c.Exec(ctx, query, args...); err == nil {
			return nil
		} else {
			return err
		}
	}
	return err
}
