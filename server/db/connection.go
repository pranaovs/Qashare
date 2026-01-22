// Package db provides database connection management and operations for the shared expenses application.
package db

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/utils"
)

// DBConfig holds database configuration parameters
type DBConfig struct {
	URL               string
	MaxConnections    int32
	MinConnections    int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

// Connect establishes a connection to the PostgreSQL database using the provided URL.
// It will attempt to create the database if it doesn't exist.
// Returns a connection pool or an error if connection fails.
func Connect(dbURL string) (*pgxpool.Pool, error) {
	config := DBConfig{
		URL:               dbURL,
		MaxConnections:    25,
		MinConnections:    2,
		MaxConnLifetime:   utils.GetEnvDuration("DB_MAX_CONN_LIFETIME", 60*60),  // 1 hour
		MaxConnIdleTime:   utils.GetEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*60), // 30 minutes
		HealthCheckPeriod: utils.GetEnvDuration("DB_HEALTH_CHECK_PERIOD", 60),   // 1 minute
		ConnectTimeout:    utils.GetEnvDuration("DB_CONNECT_TIMEOUT", 10),       // 10 seconds
	}
	return ConnectWithConfig(config)
}

// ConnectWithConfig establishes a connection to the PostgreSQL database using the provided configuration.
// It will attempt to create the database if it doesn't exist.
// Returns a connection pool or an error if connection fails.
func ConnectWithConfig(config DBConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Parse the database URL to extract database name
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")

	// Try to connect to the database
	log.Printf("[DB] Attempting to connect to database: %s", dbName)
	pool, err := createPool(ctx, config)
	if err != nil {
		// Check if error is due to database not existing
		if strings.Contains(err.Error(), "database") && strings.Contains(err.Error(), "does not exist") {
			log.Printf("[DB] Database '%s' does not exist, attempting to create it", dbName)

			// Try to create the database
			if createErr := createDatabase(config.URL, dbName); createErr != nil {
				return nil, fmt.Errorf("failed to create database: %w", createErr)
			}

			// Retry connection after creating database
			log.Printf("[DB] Retrying connection after database creation")
			pool, err = createPool(ctx, config)
			if err != nil {
				return nil, fmt.Errorf("failed to connect after database creation: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	// Verify connection with ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("[DB] Successfully connected to database: %s", dbName)
	return pool, nil
}

// createPool creates a new connection pool with the provided configuration
func createPool(ctx context.Context, config DBConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Apply configuration
	poolConfig.MaxConns = config.MaxConnections
	poolConfig.MinConns = config.MinConnections
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	return pgxpool.NewWithConfig(ctx, poolConfig)
}

// createDatabase attempts to create a new database by connecting to the 'postgres' maintenance database
func createDatabase(dbURL, dbName string) error {
	// Parse URL and replace database name with 'postgres' to connect to maintenance DB
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	parsedURL.Path = "/postgres"
	maintenanceURL := parsedURL.String()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Connect to postgres database
	conn, err := pgx.Connect(ctx, maintenanceURL)
	if err != nil {
		return fmt.Errorf("failed to connect to maintenance database: %w", err)
	}
	defer conn.Close(ctx)

	// Sanitize database name to prevent SQL injection
	// Database names must be valid PostgreSQL identifiers
	sanitizedName := sanitizeIdentifier(dbName)
	createDBSQL := fmt.Sprintf("CREATE DATABASE %s", sanitizedName)

	_, err = conn.Exec(ctx, createDBSQL)
	if err != nil {
		return fmt.Errorf("failed to execute CREATE DATABASE: %w", err)
	}

	log.Printf("[DB] Successfully created database: %s", dbName)
	return nil
}

// sanitizeIdentifier properly quotes a PostgreSQL identifier (table/database name)
// to prevent SQL injection and handle special characters
func sanitizeIdentifier(name string) string {
	// Replace any double quotes with escaped double quotes
	escaped := strings.ReplaceAll(name, `"`, `""`)
	// Quote the identifier
	return fmt.Sprintf(`"%s"`, escaped)
}

// Close gracefully closes the database connection pool
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		log.Println("[DB] Closing database connection pool")
		pool.Close()
	}
}

// HealthCheck performs a health check on the database connection
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("connection pool is nil")
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}
