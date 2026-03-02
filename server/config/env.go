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

// getEnvDuration parses a duration string with a suffix: s (seconds), m (minutes), h (hours), d (days).
// A bare number without a suffix is treated as seconds. Examples: "30s", "15m", "24h", "7d", "3600".
// Non-positive values (zero or negative) are rejected and the default is used instead.
func getEnvDuration(key string, defaultValue string) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		valStr = defaultValue
	}

	d, err := parseDuration(valStr)
	if err != nil {
		slog.Warn("Invalid duration config value", "key", key, "value", valStr, "default", defaultValue)

		// Try to parse the default value
		defaultDuration, defaultErr := parseDuration(defaultValue)
		if defaultErr != nil {
			panic("invalid default duration for " + key + ": " + defaultValue)
		}
		return defaultDuration
	}
	return d
}

func parseDuration(s string) (time.Duration, error) {
	if len(s) == 0 {
		return 0, strconv.ErrSyntax
	}

	suffix := s[len(s)-1]
	switch suffix {
	case 's', 'm', 'h':
		d, err := time.ParseDuration(s)
		if err != nil || d <= 0 {
			return 0, strconv.ErrSyntax
		}
		return d, nil
	case 'd':
		val, err := strconv.Atoi(s[:len(s)-1])
		if err != nil || val <= 0 {
			return 0, strconv.ErrSyntax
		}
		return time.Duration(val) * 24 * time.Hour, nil
	default:
		val, err := strconv.Atoi(s)
		if err != nil || val <= 0 {
			return 0, strconv.ErrSyntax
		}
		return time.Duration(val) * time.Second, nil
	}
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
