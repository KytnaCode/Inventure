package logging

import "log/slog"

// Logging attributes keys.
const (
	// KeyError is the key used for error log attributes.
	KeyError = "error"

	// KeyHandler is the key used for HTTP handler log attributed.
	KeyHandler = "handler"
)

// Error creates a new [slog.Attr] with [KeyError] as key and error as value
func Error(err error) slog.Attr {
	return slog.Any(KeyError, err)
}

// Handler creates a new [slog.Attr] with [KeyHandler] as key and handler's name as value.
func Handler(name string) slog.Attr {
	return slog.String(KeyHandler, name)
}
