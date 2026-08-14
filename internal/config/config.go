package config

// Config represents app's configuration.
type Config struct {
	// Debug is true if the app is on debug mode.
	Debug bool

	API API
}

// API contains API configuration options.
type API struct {
	// TrustedProxies is a list of trusted CIDR prefixes, if none the application will assume
	// is directly exposed to internet.
	TrustedProxies []string
}
