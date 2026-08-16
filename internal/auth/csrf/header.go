package csrf

import "net/http"

const HeaderCSRF = "X-CSRF-Token"

func FromHeader(r *http.Request) string {
	return r.Header.Get(HeaderCSRF)
}
