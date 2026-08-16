package web

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// IPMiddleware extracts client's IP from request, can be get with [middleware.GetClientIP]. If
// trustedProxies is empty, application will assume that is directly exposed to internet, and will
// use connection's remote address as client IP, if trustedProxies is not empty, it will extract
// client's IP from X-Forwarded-For.
func IPMiddleware(trustedProxies []string) func(next http.Handler) http.Handler {
	if len(trustedProxies) == 0 {
		return middleware.ClientIPFromRemoteAddr
	}

	return middleware.ClientIPFromXFF(trustedProxies...)
}
