// Package dbtest contains helper to run data access layer tests with multiple databases.
package dbtest

import (
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database types.
const (
	TypeSqlite   = "sqlite"
	TypePostgres = "postgres"
)

// dbFn creates a new [gorm.DB].
type dbFn = func(t *testing.T, migrationsFn MigrationsFn) *gorm.DB

// MigrationsFn runs migrations on a [gorm.DB].
type MigrationsFn = func(db *gorm.DB) error

// NewSqliteDB creates a new SQLite database.
func NewSqliteDB(t *testing.T, migrationsFn MigrationsFn) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"))
	if err != nil {
		t.Fatalf("could not open test database: %v", err)
	}

	err = migrationsFn(db)
	if err != nil {
		t.Fatalf("failed to run migrations for %v: %v", TypeSqlite, err)
	}

	return db
}

// NewPostgresDB spawns a Postgres test instance and connects to it.
func NewPostgresDB(t *testing.T, migrationsFn MigrationsFn) *gorm.DB {
	t.Helper()

	postgresC, err := postgres.Run(
		t.Context(),
		"postgres:18-alpine",
		postgres.BasicWaitStrategies(),
	)

	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(postgresC); err != nil {
			t.Errorf("failed to terminate postgres container: %s", err)
		}
	})

	if err != nil {
		t.Fatalf("could not create postgres container: %v", err)
	}

	cns, err := postgresC.ConnectionString(t.Context())
	if err != nil {
		t.Fatalf("could not determine postgres connection string: %v", err)
	}

	db, err := gorm.Open(postgresdriver.Open(cns))
	if err != nil {
		t.Fatalf("could not connect to postgres test instance: %v", err)
	}

	err = migrationsFn(db)
	if err != nil {
		t.Fatalf("failed to run migrations for %v: %v", TypePostgres, err)
	}

	return db
}

// RunWithDatabases runs the given runFn with all supported databases.
func RunWithDatabases(
	t *testing.T,
	migrations MigrationsFn,
	runFn func(t *testing.T, data *gorm.DB),
) {
	t.Helper()

	data := map[string]dbFn{
		TypeSqlite:   NewSqliteDB,
		TypePostgres: NewPostgresDB,
	}

	for name, newDB := range data {
		t.Run(name, func(t *testing.T) {
			db := newDB(t, migrations)

			runFn(t, db)
		})
	}
}
