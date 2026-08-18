package config

import "os"

// Environment only configuration variables.
const (
	EnvAdminUser = EnvPrefix + "ADMIN_USER"
	EnvAdminPass = EnvPrefix + "ADMIN_PASS"
)

// Env reads [Environment] from environment variables.
func Env() *Environment {
	env := new(Environment)

	env.AdminUser = os.Getenv(EnvAdminUser)
	env.AdminPass = os.Getenv(EnvAdminPass)

	return env
}
