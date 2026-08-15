package config

import (
	"os"
	"strings"
)

const (
	// EnvPrefix is the prefix used for all configuration environment variables.
	EnvPrefix = "INVENTURE_"

	// EnvAPIPrefix is the prefix used for API configuration environment variables.
	EnvAPIPrefix = EnvPrefix + "API_"

	// EnvDatabasePrefix is the prefix used for database configuration environment variables.
	EnvDatabasePrefix = EnvPrefix + "DATABASE_"

	// EnvSqlitePrefix is the prefix used for sqlite specific configuration environment variables.
	EnvSqlitePrefix = EnvDatabasePrefix + "SQLITE_"
)

// Environment variable names used for configuration.
const (
	// General
	EnvDebug = EnvPrefix + "DEBUG"

	// API
	EnvTrustedProxies = EnvAPIPrefix + "TRUSTED_PROXIES"

	// Database
	EnvDatabaseType = EnvDatabasePrefix + "TYPE"

	// SQLite
	EnvSqliteConnectionString = EnvSqlitePrefix + "CONNECTION_STRING"
)

// FromEnv reads configuration from environment. Doesn't validate data, if variable is not
// set it will return the zero value for that field.
func FromEnv() *Config {
	return &Config{
		Debug: envExists(EnvDebug),
		API: API{
			TrustedProxies: strings.Split(os.Getenv(EnvTrustedProxies), ","),
		},
		Database: Database{
			Typ: os.Getenv(EnvDatabaseType),
			SQLite: SQLiteConfig{
				ConnectionString: os.Getenv(EnvSqliteConnectionString),
			},
		},
	}
}

// envExists check if a variable is set, even if it's empty.
func envExists(env string) bool {
	_, ok := os.LookupEnv(env)

	return ok
}
