package web

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/kytnacode/inventure/pkg/logging"
)

// NewLoggerMiddleware creates a new request logger middleware.
func NewLoggerMiddleware(logger *slog.Logger, debug bool) func(next http.Handler) http.Handler {
	var level slog.Level

	if debug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	isDebug := func(_ *http.Request) bool {
		return debug
	}

	return httplog.RequestLogger(logger, &httplog.Options{
		Level:           level,
		RecoverPanics:   true,
		LogRequestBody:  isDebug,
		LogResponseBody: isDebug,
		LogBodyMaxLen:   2048,
	})
}

// WithEmbeddedLogger embeds a logger into request context that can be retrieved with
// [logging.FromCtx] and injects request's ID, if any, to logger's attributes.
func WithEmbeddedLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			reqID := middleware.GetReqID(ctx)
			if reqID != "" {
				logger = logger.With(logRequestID(reqID))
			}

			ctx = logging.WithLogger(r.Context(), logger)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
