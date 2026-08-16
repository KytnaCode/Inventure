// Package routes contains application routes.
package routes

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kytnacode/inventure/internal/auth/csrf"
	authroutes "github.com/kytnacode/inventure/internal/auth/routes"
	"github.com/kytnacode/inventure/pkg/logging"
)

// Config is the main API's configuration.
type Config struct {
	LoggerMiddleware         func(next http.Handler) http.Handler
	IPMiddleware             func(next http.Handler) http.Handler
	EmbeddedLoggerMiddelware func(next http.Handler) http.Handler
	SessionManager           *scs.SessionManager
	AuthRoutes               *authroutes.Routes
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
	middlewares := make([]func(next http.Handler) http.Handler, 0, 6)

	middlewares = append(middlewares, middleware.Recoverer)
	middlewares = append(middlewares, middleware.RequestID)
	middlewares = append(middlewares, middleware.CleanPath)

	if conf.IPMiddleware != nil {
		middlewares = append(middlewares, conf.IPMiddleware)
	}

	if conf.LoggerMiddleware != nil {
		middlewares = append(middlewares, conf.LoggerMiddleware)
	}

	if conf.EmbeddedLoggerMiddelware != nil {
		middlewares = append(middlewares, conf.EmbeddedLoggerMiddelware)
	}

	return middlewares
}

// SetupRouter setups API's root router.
func SetupRouter(conf *Config) http.Handler {
	r := chi.NewRouter()

	r.Use(baseMiddlewares(conf)...)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", HealthCheck)
		r.Get("/csrf", csrf.HandleCSRF(conf.SessionManager))

		r.Mount("/auth", conf.AuthRoutes.SetupRouter())
	})

	return r
}
