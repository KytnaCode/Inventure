package logging

import (
	"log/slog"
	"os"
)

// New returns a new JSON logger that outputs to standard output.
func New(debug bool) *slog.Logger {
	var lvl slog.Level

	if debug {
		lvl = slog.LevelDebug
	} else {
		lvl = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: debug,
		Level:     lvl,
	}))
}
