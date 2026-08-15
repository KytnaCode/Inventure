package api

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// KeyIP extracts IP address from request context.
func KeyIP(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}
