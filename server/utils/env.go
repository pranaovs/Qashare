package utils

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func Loadenv() {
	// _ = godotenv.Load(GetEnv("DEFAULT_ENV_PATH", ".env.default")) // Defaults
	_ = godotenv.Load(GetEnv("ENV_PATH", ".env"))
}

// GetEnv retrieves a string, returning default if empty
func GetEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// GetEnvRequired retrieves a string, exits if missing
func GetEnvRequired(key string) string {
	val := os.Getenv(key)
	if val == "" {
		slog.Error("Required environment variable is missing", "key", key)
		os.Exit(1)
	}
	return val
}

// GetEnvInt retrieves an integer environment variable with a default value
func GetEnvInt(key string, defaultValue int) int {
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

// GetEnvIntRequired retrieves an integer environment variable, exits if missing or invalid
func GetEnvIntRequired(key string) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		slog.Error("Required environment variable is missing", "key", key)
		os.Exit(1)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		slog.Error("Environment variable must be a valid integer", "key", key)
		os.Exit(1)
	}
	return val
}

// GetEnvBool retrieves a boolean (true, 1)
func GetEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	// ParseBool handles "1", "t", "T", "true", "TRUE", "True"
	b, err := strconv.ParseBool(val)
	if err != nil {
		slog.Warn("Invalid boolean config value, using default", "key", key, "value", val, "default", defaultValue)
		return defaultValue
	}
	return b
}

// GetEnvBoolRequired retrieves a boolean, exits if missing or invalid
func GetEnvBoolRequired(key string) bool {
	val := os.Getenv(key)
	if val == "" {
		slog.Error("Required environment variable is missing", "key", key)
		os.Exit(1)
	}
	// ParseBool handles "1", "t", "T", "true", "TRUE", "True"
	b, err := strconv.ParseBool(val)
	if err != nil {
		slog.Error("Invalid boolean config value", "key", key, "value", val)
		os.Exit(1)
	}
	return b
}

// GetEnvPort retrieves an int, validating it is a valid port (1-65535) or 0
func GetEnvPort(key string, defaultValue int) int {
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

// GetEnvPortRequired retrieves a port number, exits if missing or invalid
func GetEnvPortRequired(key string) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		slog.Error("Required environment variable is missing", "key", key)
		os.Exit(1)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		slog.Error("Config port must be a number", "key", key)
		os.Exit(1)
	}

	if val < 0 || val > 65535 {
		slog.Error("Config port must be between 0 and 65535", "key", key, "value", val)
		os.Exit(1)
	}
	return val
}

// GetEnvDuration parses a string (e.g. "60") into seconds
func GetEnvDuration(key string, defaultSeconds int) time.Duration {
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

// GetEnvDurationRequired parses a string (e.g. "60") into seconds, exits if missing or invalid
func GetEnvDurationRequired(key string) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		slog.Error("Required environment variable is missing", "key", key)
		os.Exit(1)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil || val < 0 {
		slog.Error("Config duration must be valid seconds", "key", key)
		os.Exit(1)
	}
	return time.Duration(val) * time.Second
}
