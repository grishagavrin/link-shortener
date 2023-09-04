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

type InstanceStruct struct {
	Repository istorage.Repository
	SQLDB      *pgxpool.Pool
}

func Instance(l *zap.Logger, chBatch chan istorage.BatchDelete) (*InstanceStruct, error) {
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
		stor, err := ramstorage.New(l, chBatch)
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
