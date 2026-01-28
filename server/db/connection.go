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
	RetryAttempts     int
	RetryInterval     time.Duration
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
		RetryAttempts:     utils.GetEnvInt("DB_RETRY_ATTEMPTS", 5),              // 5 attempts
		RetryInterval:     utils.GetEnvDuration("DB_RETRY_INTERVAL", 5),         // 5 seconds
	}
	return ConnectWithConfig(config)
}

// ConnectWithConfig establishes a connection to the PostgreSQL database using the provided configuration.
// It will attempt to create the database if it doesn't exist.
// Returns a connection pool or an error if connection fails.
func ConnectWithConfig(config DBConfig) (*pgxpool.Pool, error) {
	// Parse the database URL to extract database name
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	log.Printf("[DB] Attempting to connect to database: %s", dbName)

	var pool *pgxpool.Pool
	var lastErr error

	// Retry connection attempts
	for attempt := 1; attempt <= config.RetryAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)

		log.Printf("[DB] Connection attempt %d/%d", attempt, config.RetryAttempts)

		pool, err = createPool(ctx, config)
		if err != nil {
			lastErr = fmt.Errorf("failed to create connection pool: %w", err)
			cancel()

			if attempt < config.RetryAttempts {
				log.Printf("[DB] Connection attempt %d failed: %v, retrying in %v", attempt, err, config.RetryInterval)
				time.Sleep(config.RetryInterval)
				continue
			}
			break
		}

		// Verify database connectivity and existence
		if err := VerifyDatabase(ctx, pool, dbName); err != nil {
			pool.Close()
			lastErr = err
			cancel()

			if attempt < config.RetryAttempts {
				log.Printf("[DB] Database verification failed on attempt %d: %v, retrying in %v", attempt, err, config.RetryInterval)
				time.Sleep(config.RetryInterval)
				continue
			}
			break
		}

		cancel()
		log.Printf("[DB] Successfully connected to database: %s on attempt %d", dbName, attempt)
		return pool, nil
	}

	return nil, fmt.Errorf("failed to connect after %d attempts, last error: %w", config.RetryAttempts, lastErr)
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
