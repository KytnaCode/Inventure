package csrf

import (
	"crypto/subtle"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/kytnacode/inventure/internal/web"
	"github.com/kytnacode/inventure/pkg/api"
	"github.com/kytnacode/inventure/pkg/logging"
)

// KeyCSRFToken is the session key for CSRF tokens.
const KeyCSRFToken = "csrf-token"

// RequireCSRF reads a CSRF token from session and compare it with a token in
// [HeaderCSRF]. If session or header token are missing or tokens mismatch then a response
// with status [http.StatusForbidden] is returned.
//
// For missing token in session the error code will be [web.CodeMissingCSRFTokenSession], for
// mismatched tokens will be [web.CodeWrongCSRFToken].
func RequireCSRF(m *scs.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logging.FromCtx(r.Context())

			logger = logger.With(logging.Middleware("auth/csrf.RequireCSRF"))

			expectedTok := m.GetString(r.Context(), KeyCSRFToken)
			if expectedTok == "" {
				logger.Warn("missing csrf token in session")

				w.WriteHeader(http.StatusForbidden)

				err := api.WriteError(
					w,
					"missing csrf token in session",
					web.CodeMissingCSRFTokenSession,
					nil,
				)
				if err != nil {
					logger.Error("could not send error response", logging.Error(err))
				}

				return
			}

			actualTok := FromHeader(r)
			if actualTok == "" {
				logger.Warn("missing csrf token in header")

				w.WriteHeader(http.StatusForbidden)

				err := api.WriteError(w, "missing csrf token header", nil, nil)
				if err != nil {
					logger.Error("could not send error response", logging.Error(err))
				}

				return
			}

			if subtle.ConstantTimeCompare([]byte(actualTok), []byte(expectedTok)) != 1 {
				logger.Warn("mismatching csrf token")

				w.WriteHeader(http.StatusForbidden)

				err := api.WriteError(w, "mismatch csrf token", web.CodeWrongCSRFToken, nil)
				if err != nil {
					logger.Error("could not send error response", logging.Error(err))
				}

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InjectToken embeds a new CSRF token in request's session if doesn't already exists and returns
// it.
func InjectToken(m *scs.SessionManager, r *http.Request) string {
	if m.Exists(r.Context(), KeyCSRFToken) {
		return m.GetString(r.Context(), KeyCSRFToken)
	}

	tok := generateToken()

	m.Put(r.Context(), KeyCSRFToken, tok)

	return tok
}
