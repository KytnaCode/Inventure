package auth

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

const KeySessionData = "session-data"

type Session struct {
	ID string
}

func init() {
	gob.Register(&Session{})
}

// SessionStoreConfig is the configuration for [NewManager].
type SessionStoreConfig struct {
	Domain      string
	IdleTimeout time.Duration
}

// NewManager creates a new [scs.SessionManager].
func NewManager(conf *SessionStoreConfig) *scs.SessionManager {
	m := scs.New()
	m.Cookie.HttpOnly = true
	m.Cookie.Domain = conf.Domain
	m.Cookie.Path = "/"
	m.Cookie.SameSite = http.SameSiteStrictMode
	m.Cookie.Secure = true
	m.IdleTimeout = conf.IdleTimeout

	return m
}
