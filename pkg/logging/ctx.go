package logging

import (
	"context"
	"log/slog"
)

var discardLogger = slog.New(slog.DiscardHandler)

type ctxLoggerKey struct{}

// WithLogger returns a new context with the given logger embedded, use [FromCtx] to get
// logger from the new context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

// FromCtx returns an embedded logger from a context, if context has no logger then a no-op logger
// will be returned.
func FromCtx(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(ctxLoggerKey{}).(*slog.Logger)
	if !ok || logger == nil {
		return discardLogger
	}

	return logger
}
