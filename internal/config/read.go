package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Read reads configuration from a configuration file or environment.
func Read() (*Config, error) {
	viper.SetConfigName("config")

	viper.SetEnvPrefix("inventure")

	viper.AddConfigPath("/etc/inventure")
	viper.AddConfigPath("$XDG_CONFIG_HOME/inventure")
	viper.AddConfigPath("$HOME/.config/inventure")
	viper.AddConfigPath("$HOME/.inventure")
	viper.AddConfigPath(".")

	viper.SetDefault("debug", DefaultDebug)
	viper.SetDefault("addr", DefaultAddr)

	viper.SetDefault("api.trusted_proxies", []string{})
	viper.SetDefault("api.password_auth.request_limit", DefaultPasswordAuthRequestLimit)
	viper.SetDefault("api.password_auth.time_window_seconds", DefaultPasswordAuthTimeWindowSeconds)
	viper.SetDefault("api.login_attempt.request_limit", DefaultLoginAttemptLimit)
	viper.SetDefault("api.login_attempt.time_window_seconds", DefaultLoginAttemptTimeWindowSeconds)

	viper.SetDefault("database.type", DefaultDBType)
	viper.SetDefault("database.sqlite.connection_string", DefaultSqliteConnectionString)

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("could not read configuration: %w", err)
	}

	c := new(Config)

	err = viper.UnmarshalExact(c)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal configuration: %w", err)
	}

	return c, nil
}
