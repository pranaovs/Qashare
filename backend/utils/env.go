package utils

import (
	"os"

	"github.com/joho/godotenv"
)

func Loadenv() {
	_ = godotenv.Load(Getenv("ENV_PATH", ".env"))
	_ = godotenv.Overload(Getenv("DEFAULT_ENV_PATH", ".env.default"))
}

// Getenv Loadenv before calling
func Getenv(key, defaultVal string) string {
	val, exists := os.LookupEnv(key)

	if !exists {
		return defaultVal
	}

	return val
}
