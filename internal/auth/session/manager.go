package session

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

// ManagerConfig is the configuration for [NewManager].
type ManagerConfig struct {
	Domain      string
	IdleTimeout time.Duration
}

// NewManager creates a new [scs.SessionManager].
func NewManager(conf *ManagerConfig) *scs.SessionManager {
	m := scs.New()
	m.Cookie.HttpOnly = true
	m.Cookie.Domain = conf.Domain
	m.Cookie.Path = "/"
	m.Cookie.SameSite = http.SameSiteStrictMode
	m.Cookie.Secure = true
	m.IdleTimeout = conf.IdleTimeout

	return m
}
