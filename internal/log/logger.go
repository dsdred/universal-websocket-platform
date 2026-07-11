package log

import (
	"io"
	"log/slog"
)

// New creates the application logger.
func New(output io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, nil))
}
