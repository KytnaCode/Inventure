// Package server handles HTTP server context-based cancellation.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Config is the configuration for [ListenAndServe].
type Config struct {
	Server          *http.Server
	ShutdownTimeout time.Duration
}

// ListenAndServe starts listening with the given server from [Config] and handles
// context-based cancellation.
func ListenAndServe(ctx context.Context, conf *Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := withOptionalTimeout(ctx, conf.ShutdownTimeout)
		defer cancel()

		errCh <- conf.Server.Shutdown(shutdownCtx)
	}()

	err := conf.Server.ListenAndServe()

	return errors.Join(err, <-errCh)
}

// withOptionalTimeout returns a context with timeout. If timeout is zero then fallback to
// [context.WithCancel].
func withOptionalTimeout(
	parent context.Context,
	timeout time.Duration,
) (ctx context.Context, cancel func()) {
	if timeout == 0 {
		return context.WithCancel(parent)
	}

	return context.WithTimeout(parent, timeout)
}
