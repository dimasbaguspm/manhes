package log

import (
	"log/slog"
	"os"
)

// New returns a structured logger at the given level ("debug", "info", "warn", "error").
func New(level string) *slog.Logger {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		l = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: l}))
	slog.SetDefault(logger)
	return logger
}
