package csrf

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
)

// HandleCSRF injects a new token if not exists in user's session and returns the token in the
// header [HeaderCSRF].
func HandleCSRF(m *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := InjectToken(m, r)

		w.Header().Set(HeaderCSRF, tok)
	}
}
