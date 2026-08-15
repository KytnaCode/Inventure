package config

// Config represents app's configuration.
type Config struct {
	// Debug is true if the app is on debug mode.
	Debug bool

	API API

	Database Database
}

// API contains API configuration options.
type API struct {
	// TrustedProxies is a list of trusted CIDR prefixes, if none the application will assume
	// is directly exposed to internet.
	TrustedProxies []string
}

// DatabaseSQLite is database type string for sqlite.
const DatabaseSQLite = "sqlite"

// Database contains database configuration.
type Database struct {
	Typ string

	SQLite SQLiteConfig
}

// SQLiteConfig contains sqlite specific configuration.
type SQLiteConfig struct {
	ConnectionString string
}
