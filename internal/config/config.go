package config

// Configuration defaults.
const (
	DefaultPasswordAuthRequestLimit      = 15
	DefaultPasswordAuthTimeWindowSeconds = 60
	DefaultLoginAttemptLimit             = 10
	DefaultLoginAttemptTimeWindowSeconds = 300
)

// Config represents app's configuration.
type Config struct {
	// Debug is true if the app is on debug mode.
	Debug bool

	// Addr is the address for the app to listen on.
	Addr string

	API API

	Database Database
}

// API contains API configuration options.
type API struct {
	// TrustedProxies is a list of trusted CIDR prefixes, if none the application will assume
	// is directly exposed to internet.
	TrustedProxies []string

	// DisableRateLimit disables included rate limit, set if using a reverse proxy for rate limiting.
	DisableRateLimit bool

	// PasswordAuthRequestLimit request limit for password-based authentication, independent from
	// email address. Set this or [PasswordAuthTimeWindowSeconds] to 0 to disable raw rate
	// limiting for authentication endpoints, alternatively set [DisableRateLimit] to disable rate
	// limiting globally.
	PasswordAuthRequestLimit int

	// PasswordAuthTimeWindowSeconds is the time window in seconds for that
	// [PasswordAuthRequestLimit] applies. Set this or [PasswordAuthRequestLimit] to 0 to disable raw
	// rate limiting for authentication endpoints, alternatively set [DisableRateLimit] to disable
	// rate limiting globally.
	PasswordAuthTimeWindowSeconds int

	// LoginAttemptLimit is the limit of login or sign up attempts for a specific email address.
	// set this or [LoginAttemptTimeWindowSeconds] to 0 to disable login attempt limit, alternatively
	// set [DisableRateLimit] to disable rate limiting globally.
	LoginAttemptLimit int

	// LoginAttemptTimeWindowSeconds is the time window in seconds for that [LoginAttemptLimit]
	// applies. Set this or [LoginAttemptLimit] to 0 to disable login attempt limit, alternatively
	// set [DisableRateLimit] to disable rate limiting globally.
	LoginAttemptTimeWindowSeconds int
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
