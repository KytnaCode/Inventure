// Package middleware implements authentication and authorization middleware.
package middleware

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/kytnacode/inventure/internal/auth/session"
	"github.com/kytnacode/inventure/pkg/api"
	"github.com/kytnacode/inventure/pkg/logging"
)

// RequireAuth checks if an authenticated session exists in request's context, if not, send an
// unauthorized error.
func RequireAuth(m *scs.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logging.FromCtx(r.Context())

			logger = logger.With(logging.Middleware("auth/middleware.RequireAuth"))

			data, ok := m.Get(r.Context(), session.KeySessionData).(*session.Session)
			if !ok || data == nil || data.ID == "" {
				logger.Info("unauthorized request")

				w.WriteHeader(http.StatusUnauthorized)

				err := api.WriteError(w, "unauthorized", nil, nil)
				if err != nil {
					logger.Error("could not write error response", logging.Error(err))
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
