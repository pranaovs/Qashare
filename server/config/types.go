package config

import "time"

// Config holds all application configuration
type Config struct {
	API      APIConfig
	Database DatabaseConfig
	JWT      JWTConfig
	App      AppConfig
}

// APIConfig holds API server configuration
type APIConfig struct {
	BasePath  string
	PublicURL string
	BindAddr  string
	BindPort  int
}

// DatabaseConfig holds database connection and pool configuration
type DatabaseConfig struct {
	URL               string
	MigrationsDir     string
	VerifyMigrations  bool
	MaxConnections    int
	MinConnections    int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
	RetryAttempts     int
	RetryInterval     time.Duration
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

// AppConfig holds general application configuration
type AppConfig struct {
	Debug          bool
	DisableSwagger bool
	SplitTolerance float64
	EnvPath        string
}
