package config

import "os"

// EnvPrefix is the prefix used for all configuration environment variables.
const EnvPrefix = "INVENTURE_"

// Environment variable names used for configuration.
const (
	EnvDebug = EnvPrefix + "DEBUG"
)

// FromEnv reads configuration from environment. Doesn't validate data, if variable is not
// set it will return the zero value for that field.
func FromEnv() *Config {
	return &Config{
		Debug: envExists(EnvDebug),
	}
}

// envExists check if a variable is set, even if it's empty.
func envExists(env string) bool {
	_, ok := os.LookupEnv(env)

	return ok
}
