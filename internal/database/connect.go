// Package database handles database connection logic.
package database

import (
	"context"
	"errors"

	"github.com/kytnacode/inventure/internal/config"
	"github.com/kytnacode/inventure/pkg/logging"
	"github.com/kytnacode/inventure/pkg/retry"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Connect connects to a database instance, database type is selected via configuration. Log
// errors automatically using [logging.FromCtx]. If successfully return the database instance
// and `ok` set to true.
//
// If ok is false, the connection failed, errors were logged and a nil
// database was returned.
func Connect(ctx context.Context, conf *config.Database) (db *gorm.DB, ok bool) {
	logger := logging.FromCtx(ctx)

	typ := conf.Typ

	if typ == "" {
		logger.Warn("no database type selected: fallback to sqlite")

		typ = config.DatabaseSQLite
	}

	if typ == config.DatabaseSQLite {
		return connectSQLite(ctx, conf.SQLite)
	}

	logger.Error("unknown database type")

	return nil, false
}

func connectSQLite(ctx context.Context, conf *config.SQLiteConfig) (db *gorm.DB, ok bool) {
	logger := logging.FromCtx(ctx)

	err := retry.Do(ctx, func(_ context.Context) (temporary bool, err error) {
		gormDatabase, err := gorm.Open(sqlite.Open(conf.ConnectionString))
		if err == nil {
			db = gormDatabase

			return false, nil
		}

		if errors.Is(err, gorm.ErrUnsupportedDriver) {
			return false, err
		}

		return true, err
	})
	if err != nil {
		logger.Error("could not open sqlite database", logging.Error(err))

		return nil, false
	}

	return db, true
}
