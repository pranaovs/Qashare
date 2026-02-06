package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	JwtRandomSecretLength = 32
)

// Load reads environment variables and returns a populated Config struct
// It fails fast if required configuration is missing or invalid
func Load() (*Config, error) {
	// Load .env file
	envPath := getEnv("ENV_PATH", ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("[WARNING] Could not load .env file at %s: %v", envPath, err)
	}

	cfg := &Config{}

	// Load API configuration
	cfg.API = loadAPIConfig()

	// Load Database configuration
	cfg.Database = loadDatabaseConfig()

	// Load JWT configuration
	cfg.JWT = loadJWTConfig()

	// Load App configuration
	cfg.App = loadAppConfig(envPath)

	log.Println("[CONFIG] Configuration loaded successfully")
	return cfg, nil
}

func loadAPIConfig() APIConfig {
	return APIConfig{
		BasePath:  getEnv("API_BASE_PATH", "/api"),
		PublicURL: getEnv("API_PUBLIC_URL", "http://localhost:8080"),
		BindAddr:  getEnv("API_BIND_ADDR", "0.0.0.0"),
		BindPort:  getEnvPort("API_BIND_PORT", 8080),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		URL:               getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/qashare"),
		MigrationsDir:     getEnv("DB_MIGRATIONS_DIR", "migrations"),
		VerifyMigrations:  getEnvBool("DB_VERIFY_MIGRATIONS", true),
		MaxConnections:    getEnvInt("DB_MAX_CONNECTIONS", 10),
		MinConnections:    getEnvInt("DB_MIN_CONNECTIONS", 2),
		MaxConnLifetime:   getEnvDuration("DB_MAX_CONN_LIFETIME", 60*60),
		MaxConnIdleTime:   getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*60),
		HealthCheckPeriod: getEnvDuration("DB_HEALTH_CHECK_PERIOD", 60),
		ConnectTimeout:    getEnvDuration("DB_CONNECT_TIMEOUT", 10),
		RetryAttempts:     getEnvInt("DB_RETRY_ATTEMPTS", 5),
		RetryInterval:     getEnvDuration("DB_RETRY_INTERVAL", 5),
	}
}

func loadJWTConfig() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Printf("[WARNING] JWT_SECRET not provided, using random value. Tokens will not be remembered across restarts.")
		secret = generateRandomSecret(JwtRandomSecretLength)
	}

	return JWTConfig{
		Secret: secret,
		Expiry: getEnvDuration("JWT_EXPIRY", 60*60*24), // 24 hours
	}
}

func loadAppConfig(envPath string) AppConfig {
	return AppConfig{
		Debug:          getEnvBool("DEBUG", false),
		DisableSwagger: getEnvBool("DISABLE_SWAGGER", false),
		SplitTolerance: getEnvFloat("SPLIT_TOLERANCE", 0.01),
		EnvPath:        envPath,
	}
}

func generateRandomSecret(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal("failed to generate random bytes for JWT secret:", err)
	}

	return base64.StdEncoding.EncodeToString(b)
}
