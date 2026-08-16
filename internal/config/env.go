package config

import (
	"os"
	"strconv"
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
	EnvTrustedProxies                = EnvAPIPrefix + "TRUSTED_PROXIES"
	EnvDisableRateLimit              = EnvAPIPrefix + "RATE_LIMIT_DISABLE"
	EnvPasswordAuthRequestLimit      = EnvAPIPrefix + "PASSWORD_AUTH_REQUEST_LIMIT"
	EnvPasswordAuthTimeWindowSeconds = EnvAPIPrefix + "PASSWORD_AUTH_TIME_WINDOW_SECONDS"
	EnvLoginAttemptLimit             = EnvAPIPrefix + "LOGIN_ATTEMPT_LIMIT"
	EnvLoginAttemptTimeWindowSeconds = EnvAPIPrefix + "LOGIN_ATTEMPT_TIME_WINDOW_SECONDS"

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
			TrustedProxies:   strings.Split(os.Getenv(EnvTrustedProxies), ","),
			DisableRateLimit: envExists(EnvDisableRateLimit),
			PasswordAuthRequestLimit: intOr(
				EnvPasswordAuthRequestLimit,
				DefaultPasswordAuthRequestLimit,
			),
			PasswordAuthTimeWindowSeconds: intOr(
				EnvPasswordAuthTimeWindowSeconds,
				DefaultPasswordAuthTimeWindowSeconds,
			),
			LoginAttemptLimit: intOr(
				EnvLoginAttemptLimit,
				DefaultLoginAttemptLimit,
			),
			LoginAttemptTimeWindowSeconds: intOr(
				EnvLoginAttemptTimeWindowSeconds,
				DefaultLoginAttemptTimeWindowSeconds,
			),
		},
		Database: Database{
			Typ: os.Getenv(EnvDatabaseType),
			SQLite: SQLiteConfig{
				ConnectionString: os.Getenv(EnvSqliteConnectionString),
			},
		},
	}
}

// envExists check if a variable is set, if set to an empty string or 0 returns false.
func envExists(env string) bool {
	v, ok := os.LookupEnv(env)
	if ok && (v == "" || v == "0") {
		return false
	}

	return ok
}

func intOr(env string, def int) int {
	raw, ok := os.LookupEnv(env)
	if !ok {
		return def
	}

	v, err := strconv.ParseInt(raw, 10, 0)
	if err != nil {
		return 0
	}

	return int(v)
}
