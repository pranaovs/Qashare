package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Helper functions for reading environment variables
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		slog.Warn("Invalid integer config value, using default", "key", key, "value", valStr, "default", defaultValue)
		return defaultValue
	}
	return val
}

func getEnvInt32(key string, defaultValue int32) int32 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.ParseInt(valStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid integer config value, using default", "key", key, "value", valStr, "default", defaultValue)
		return defaultValue
	}
	return int32(val)
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		slog.Warn("Invalid boolean config value, using default", "key", key, "value", val, "default", defaultValue)
		return defaultValue
	}
	return b
}

func getEnvFloat(key string, defaultValue float64) float64 {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		slog.Warn("Invalid float config value, using default", "key", key, "value", valStr, "default", defaultValue)
		return defaultValue
	}
	return val
}

func getEnvPort(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		slog.Warn("Config port must be a number, using default", "key", key, "default", defaultValue)
		return defaultValue
	}

	if val < 0 || val > 65535 {
		slog.Warn("Config port out of range, using default", "key", key, "value", val, "default", defaultValue)
		return defaultValue
	}
	return val
}

func getEnvDuration(key string, defaultSeconds int) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return time.Duration(defaultSeconds) * time.Second
	}

	val, err := strconv.Atoi(valStr)
	if err != nil || val < 0 {
		slog.Warn("Config duration must be valid seconds, using default", "key", key, "default", defaultSeconds)
		return time.Duration(defaultSeconds) * time.Second
	}
	return time.Duration(val) * time.Second
}

func getEnvList(key string, defaultVal []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
