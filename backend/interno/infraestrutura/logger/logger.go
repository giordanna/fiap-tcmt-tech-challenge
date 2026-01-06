package logger

import (
	"log/slog"
	"os"
)

// InitLogger configura o logger para JSON (melhor para Cloud Logging)
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// handler json para logs estruturados
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	slog.SetDefault(logger)
}