package utils

import (
	"log/slog"
	"os"
)

// Logger тонкая обертка над slog для единообразия.
type Logger struct {
	*slog.Logger
}

// NewLogger создает JSON-логгер; более подробный вывод в dev-среде.
func NewLogger(env string) *Logger {
	level := slog.LevelInfo
	if env == "dev" || env == "local" {
		level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return &Logger{Logger: slog.New(handler)}
}
