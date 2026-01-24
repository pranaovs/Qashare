// Package db provides database connection management and operations for the shared expenses application.
package db

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

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
		MaxConnections:    int32(utils.GetEnvInt("DB_MAX_CONNECTIONS", 10)),
		MinConnections:    int32(utils.GetEnvInt("DB_MIN_CONNECTIONS", 2)),
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
	log.Printf("[DB] Attempting to connect to database: %s", dbName)

	pool, err := createPool(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify database connectivity and existence
	if err := VerifyDatabase(ctx, pool, dbName); err != nil {
		pool.Close()
		return nil, err
	}

	log.Printf("[DB] Successfully connected to database: %s", dbName)
	return pool, nil
}

func VerifyDatabase(ctx context.Context, pool *pgxpool.Pool, dbName string) error {
	if pool == nil {
		return fmt.Errorf("connection pool is nil")
	}

	if err := pool.Ping(ctx); err != nil {
		// pgx wraps server errors; keep message intact for operators
		return fmt.Errorf(
			"database verification failed for '%s': %w",
			dbName,
			err,
		)
	}

	return nil
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
