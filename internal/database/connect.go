// Package database handles database connection logic.
package database

import (
	"context"
	"errors"

	"github.com/kytnacode/inventure/internal/config"
	"github.com/kytnacode/inventure/retry"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ErrUnsupportedDBType is returned when database type is set to an unsupported database.
var ErrUnsupportedDBType = errors.New("unsupported DB type")

// Connect connects to a database instance and return a gorm db.
func Connect(ctx context.Context, conf *config.Database) (db *gorm.DB, err error) {
	typ := conf.Typ

	if typ == config.DatabaseSQLite {
		return connectSQLite(ctx, &conf.SQLite)
	}

	return nil, ErrUnsupportedDBType
}

func connectSQLite(ctx context.Context, conf *config.SQLiteConfig) (db *gorm.DB, err error) {
	err = retry.Do(ctx, func(_ context.Context) (temporary bool, err error) {
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
		return nil, err
	}

	return db, nil
}
