package utils

import (
	"context"
	"log/slog"
	"os"

	"github.com/pranaovs/qashare/config"
)

var logger = slog.Default()

// InitLogger initializes the structured logger with the provided config
func InitLogger(cfg *config.Config) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if cfg.App.Debug {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// Logger returns the global structured logger
func Logger() *slog.Logger {
	return logger
}

// LogError logs an error with context
func LogError(ctx context.Context, msg string, err error, attrs ...any) {
	allAttrs := append([]any{"error", err}, attrs...)
	logger.ErrorContext(ctx, msg, allAttrs...)
}

// LogInfo logs an informational message
func LogInfo(ctx context.Context, msg string, attrs ...any) {
	logger.InfoContext(ctx, msg, attrs...)
}

// LogDebug logs a debug message
func LogDebug(ctx context.Context, msg string, attrs ...any) {
	logger.DebugContext(ctx, msg, attrs...)
}

// LogWarn logs a warning message
func LogWarn(ctx context.Context, msg string, attrs ...any) {
	logger.WarnContext(ctx, msg, attrs...)
}
