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
)

// Environment variable names used for configuration.
const (
	EnvDebug          = EnvPrefix + "DEBUG"
	EnvTrustedProxies = EnvAPIPrefix + "TRUSTED_PROXIES"
)

// FromEnv reads configuration from environment. Doesn't validate data, if variable is not
// set it will return the zero value for that field.
func FromEnv() *Config {
	return &Config{
		Debug: envExists(EnvDebug),
		API: API{
			TrustedProxies: strings.Split(os.Getenv(EnvTrustedProxies), ","),
		},
	}
}

// envExists check if a variable is set, even if it's empty.
func envExists(env string) bool {
	_, ok := os.LookupEnv(env)

	return ok
}
