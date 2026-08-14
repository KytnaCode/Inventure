package web

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"
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
