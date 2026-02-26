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
	BasePath       string   `example:"/api"`
	PublicURL      string   `example:"http://localhost:8080"`
	BindAddr       string   `example:"0.0.0.0"`
	BindPort       int      `example:"8080"`
	TrustedProxies []string `example:"127.0.0.1,192.168.0.1"`
}

// DatabaseConfig holds database connection and pool configuration
type DatabaseConfig struct {
	URL               string        `example:"postgres://postgres:postgres@localhost:5432/qashare"`
	MigrationsDir     string        `example:"migrations"`
	VerifyMigrations  bool          `example:"true"`
	MaxConnections    int32         `example:"10"`
	MinConnections    int32         `example:"2"`
	MaxConnLifetime   time.Duration `example:"1h"`
	MaxConnIdleTime   time.Duration `example:"30m"`
	HealthCheckPeriod time.Duration `example:"60s"`
	ConnectTimeout    time.Duration `example:"10s"`
	RetryAttempts     int           `example:"5"`
	RetryInterval     time.Duration `example:"5s"`
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret        string        `example:"random-generated-secret"`
	Audience      string        `example:"qashare"`
	Issuer        string        `example:"qashare"`
	RefreshExpiry time.Duration `example:"30d"`
	AccessExpiry  time.Duration `example:"15m"`
}

// AppConfig holds general application configuration
type AppConfig struct {
	Debug          bool    `example:"false"`
	DisableSwagger bool    `example:"false"`
	SplitTolerance float64 `example:"0.01"`
	EnvPath        string  `example:".env"`
}
