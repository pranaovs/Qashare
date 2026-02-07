package config

import (
	"log"
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
		log.Printf("Config Warning: Invalid integer for %s: '%s', using default: %d", key, valStr, defaultValue)
		return defaultValue
	}
	return val
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Printf("Config Warning: Invalid boolean for %s: '%s', using default: %t", key, val, defaultValue)
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
		log.Printf("Config Warning: Invalid float for %s: '%s', using default: %f", key, valStr, defaultValue)
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
		log.Printf("Config Warning: %s must be a number, using default %d", key, defaultValue)
		return defaultValue
	}

	if val < 0 || val > 65535 {
		log.Printf("Config Warning: %s must be between 0 and 65535, using default %d", key, defaultValue)
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
		log.Printf("Config Warning: %s must be a valid number of seconds, using default %d", key, defaultSeconds)
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
