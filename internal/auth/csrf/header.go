package csrf

import "net/http"

// HeaderCSRF is the header from to read CSRF tokens from client.
const HeaderCSRF = "X-Csrf-Token"

// FromHeader extracts a CSRF token from a request.
func FromHeader(r *http.Request) string {
	return r.Header.Get(HeaderCSRF)
}
