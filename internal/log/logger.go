package log

import (
	"io"
	"log/slog"
)

// New creates the application logger at the requested level.
func New(output io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level}))
}
