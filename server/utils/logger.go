package utils

import (
	"context"
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	// Initialize structured logger with JSON output
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	
	// Check environment for debug mode
	if GetEnvBool("DEBUG", false) {
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
