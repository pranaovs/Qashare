package utils

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func Loadenv() {
	_ = godotenv.Load(GetEnv("DEFAULT_ENV_PATH", ".env.default")) // Defaults
	_ = godotenv.Overload(GetEnv("ENV_PATH", ".env"))             // Overrides
}

// GetEnv retrieves a string, returning default if empty
func GetEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// GetEnvRequired retrieves a string, calls log.Fatal if missing
func GetEnvRequired(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Config Error: Required environment variable %s is missing", key)
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
		log.Printf("Warning: Invalid integer for %s: '%s', using default: %d", key, valStr, defaultValue)
		return defaultValue
	}
	return val
}

// GetEnvIntRequired retrieves an integer environment variable, calls log.Fatal if missing or invalid
func GetEnvIntRequired(key string) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Fatalf("Config Error: Required environment variable %s is missing", key)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Config Error: %s must be a valid integer", key)
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
		log.Printf("Warning: Invalid boolean for %s: '%s', using default: %t", key, val, defaultValue)
		return defaultValue
	}
	return b
}

// GetEnvBoolRequired retrieves a boolean, calls log.Fatal if missing or invalid
func GetEnvBoolRequired(key string) bool {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Config Error: Required environment variable %s is missing", key)
	}
	// ParseBool handles "1", "t", "T", "true", "TRUE", "True"
	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Fatalf("Config Error: Invalid boolean for %s: '%s'", key, val)
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
		log.Fatalf("Config Error: %s must be a number", key)
	}

	if val < 0 || val > 65535 {
		log.Fatalf("Config Error: %s must be between 0 and 65535", key)
	}
	return val
}

// GetEnvPortRequired retrieves a port number, calls log.Fatal if missing or invalid
func GetEnvPortRequired(key string) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Fatalf("Config Error: Required environment variable %s is missing", key)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Config Error: %s must be a number", key)
	}

	if val < 0 || val > 65535 {
		log.Fatalf("Config Error: %s must be between 0 and 65535", key)
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
	if err != nil {
		log.Fatalf("Config Error: %s must be a valid number of seconds", key)
	}
	return time.Duration(val) * time.Second
}

// GetEnvDurationRequired parses a string (e.g. "60") into seconds, calls log.Fatal if missing or invalid
func GetEnvDurationRequired(key string) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		log.Fatalf("Config Error: Required environment variable %s is missing", key)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Fatalf("Config Error: %s must be a valid number of seconds", key)
	}
	return time.Duration(val) * time.Second
}
