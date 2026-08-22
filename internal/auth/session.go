package auth

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

// KeySessionData is the session key for session data.
const KeySessionData = "session-data"

// Session contains session data.
type Session struct {
	// ID is user's ID.
	ID string

	// RoleIDs are the IDs of user's roles.
	RoleIDs []string
}

func init() {
	// Types must be gob-registered to be able to be stored on session.
	gob.Register(&Session{})
}

// SessionStoreConfig is the configuration for [NewManager].
type SessionStoreConfig struct {
	// Domain is web domain for use in session cookie.
	Domain string

	// IdleTimeout is the inactivity timeout for a session to expire.
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
