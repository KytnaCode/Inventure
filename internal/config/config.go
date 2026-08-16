package config

// Configuration defaults.
const (
	DefaultDebug                         = false
	DefaultAddr                          = "0.0.0.0:80"
	DefaultPasswordAuthRequestLimit      = 15
	DefaultPasswordAuthTimeWindowSeconds = 60
	DefaultLoginAttemptLimit             = 10
	DefaultLoginAttemptTimeWindowSeconds = 300
	DefaultDBType                        = "sqlite"
	DefaultSqliteConnectionString        = "./app.db"
)

// Config represents app's configuration.
type Config struct {
	// Debug is true if the app is on debug mode.
	Debug bool

	Addr string

	API API

	Database Database
}

// API contains API configuration options.
type API struct {
	// TrustedProxies is a list of trusted CIDR prefixes, if none the application will assume
	// is directly exposed to internet.
	TrustedProxies []string `mapstructure:"trusted_proxies"`

	// DisableRateLimit disables included rate limit, set if using a reverse proxy for rate limiting.
	DisableRateLimit bool `mapstructure:"disable_rate_limit"`

	// PasswordAuth is raw rate limit configuration for password-based authentication.
	PasswordAuth RateLimit `mapstructure:"password_auth"`

	// LoginAttempt is rate limit configuration for specific email or login credentials.
	LoginAttempt RateLimit `mapstructure:"login_attempt"`
}

// Server contains server related configuration.
type Server struct {
	// Addr is the address for the app to listen on.
	Addr string

	// CertFile is server's SSL certificate, if both this and a key is set, server will run
	// using HTTPS.
	CertFile string `mapstructure:"ssl_cert_file"`

	// KeyFile is server's SSL certificate key, if both this and a certificate is set, server will
	// run using HTTPS.
	KeyFile string `mapstructure:"ssl_key_file"`
}

// RateLimit contains rate limit configuration for a specific endpoint or group of endpoints.
type RateLimit struct {
	// Set this or [TimeWindowSeconds] to 0 to disable rate limiting for endpoints using this,
	// configuration, alternatively set [DisableRateLimit] to disable rate limiting globally.
	RequestLimit int `mapstructure:"request_limit"`

	// TimeWindowSeconds is the time window in seconds for that [RequestLimit] applies. Set this
	// or [RequestLimit] to 0 to disable rate limiting for endpoints using this configuration,
	// alternatively set [DisableRateLimit] to disable rate limiting globally.
	TimeWindowSeconds int `mapstructure:"time_window_seconds"`
}

// DatabaseSQLite is database type string for sqlite.
const DatabaseSQLite = "sqlite"

// Database contains database configuration.
type Database struct {
	Typ string `mapstructure:"type"`

	SQLite SQLiteConfig
}

// SQLiteConfig contains sqlite specific configuration.
type SQLiteConfig struct {
	ConnectionString string `mapstructure:"connection_string"`
}
