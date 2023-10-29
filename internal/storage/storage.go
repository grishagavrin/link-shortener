// Package storage implement db connection
package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/storage/dbstorage"
	"github.com/grishagavrin/link-shortener/internal/storage/filestorage"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var instance *pgxpool.Pool

// SQLDBConnection create pgx pool connection
func SQLDBConnection(l *zap.Logger) (*pgxpool.Pool, error) {
	if instance == nil {
		// config instance
		cfg, err := config.Instance()
		if errors.Is(err, errs.ErrENVLoading) {
			return nil, errs.ErrDatabaseNotAvaliable
		}

		// dsn config value
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

// InstanceStruct instance struct for repository & pgpool connection
type InstanceStruct struct {
	Repository models.Repository
	SQLDB      *pgxpool.Pool
}

// Instance initialize storage with channel for batch delete
func Instance(l *zap.Logger, chBatch chan models.BatchDelete) (*InstanceStruct, error) {
	dbi, err := SQLDBConnection(l)
	instanceDB := &InstanceStruct{}

	if err == nil {
		stor, err := dbstorage.New(dbi, l, chBatch)
		if errors.Is(err, errs.ErrDatabaseNotAvaliable) || errors.Is(err, errs.ErrDatabaseExec) {
			instanceDB.Repository = nil
			instanceDB.SQLDB = dbi
			return instanceDB, err
		}

		// Butch delete listener for SQL database
		go stor.BunchUpdateAsDeleted(chBatch)
		l.Info("Connected to DB")
		instanceDB.Repository = stor
		instanceDB.SQLDB = dbi
		return instanceDB, nil
	} else {
		stor, err := filestorage.New(l, chBatch)
		if errors.Is(err, errs.ErrRAMNotAvaliable) {
			instanceDB.Repository = nil
			instanceDB.SQLDB = nil
			return instanceDB, err
		}

		// Butch delete listener for RAM database
		go stor.BunchUpdateAsDeleted(chBatch)
		l.Info("Set RAM handler")
		instanceDB.Repository = stor
		instanceDB.SQLDB = nil
		return instanceDB, nil
	}
}
