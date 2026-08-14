// Package routes contains application routes.
package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kytnacode/inventure/pkg/logging"
)

// Config is the main API's configuration.
type Config struct {
	LoggerMiddleware func(next http.Handler) http.Handler
	IPMiddleware     func(next http.Handler) http.Handler
}

// HealthCheck is a health check handler.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("healthy"))
	if err != nil {
		logger := logging.FromCtx(r.Context())

		logger.Error("could not send health check response", logging.Error(err))
	}
}

func baseMiddlewares(conf *Config) []func(next http.Handler) http.Handler {
	middlewares := make([]func(next http.Handler) http.Handler, 0, 5)

	middlewares = append(middlewares, middleware.Recoverer)
	middlewares = append(middlewares, middleware.RequestID)
	middlewares = append(middlewares, middleware.CleanPath)

	if conf.IPMiddleware != nil {
		middlewares = append(middlewares, conf.IPMiddleware)
	}

	if conf.LoggerMiddleware != nil {
		middlewares = append(middlewares, conf.LoggerMiddleware)
	}

	return middlewares
}

// SetupRouter setups API's root router.
func SetupRouter(conf *Config) http.Handler {
	r := chi.NewRouter()

	r.Use(baseMiddlewares(conf)...)

	r.Get("/api/v1/health", HealthCheck)

	return r
}
