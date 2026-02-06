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
	"github.com/pranaovs/qashare/config"
)

// Connect establishes a connection to the PostgreSQL database using the provided configuration.
// It will attempt to create the database if it doesn't exist.
// Returns a connection pool or an error if connection fails.
func Connect(dbConfig config.DatabaseConfig) (*pgxpool.Pool, error) {
	// Parse the database URL to extract database name
	parsedURL, err := url.Parse(dbConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	log.Printf("[DB] Attempting to connect to database: %s", dbName)

	var pool *pgxpool.Pool
	var lastErr error

	// Retry connection attempts
	for attempt := 1; attempt <= dbConfig.RetryAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), dbConfig.ConnectTimeout)

		log.Printf("[DB] Connection attempt %d/%d", attempt, dbConfig.RetryAttempts)

		pool, err = createPool(ctx, dbConfig)
		if err != nil {
			lastErr = fmt.Errorf("failed to create connection pool: %w", err)
			cancel()

			if attempt < dbConfig.RetryAttempts {
				log.Printf("[DB] Connection attempt %d failed: %v, retrying in %v", attempt, err, dbConfig.RetryInterval)
				time.Sleep(dbConfig.RetryInterval)
				continue
			}
			break
		}

		// Verify database connectivity and existence
		if err := VerifyDatabase(ctx, pool, dbName); err != nil {
			pool.Close()
			lastErr = err
			cancel()

			if attempt < dbConfig.RetryAttempts {
				log.Printf("[DB] Database verification failed on attempt %d: %v, retrying in %v", attempt, err, dbConfig.RetryInterval)
				time.Sleep(dbConfig.RetryInterval)
				continue
			}
			break
		}

		cancel()
		log.Printf("[DB] Successfully connected to database: %s on attempt %d", dbName, attempt)
		return pool, nil
	}

	return nil, fmt.Errorf("failed to connect after %d attempts, last error: %w", dbConfig.RetryAttempts, lastErr)
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
func createPool(ctx context.Context, dbConfig config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dbConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Apply configuration
	poolConfig.MaxConns = int32(dbConfig.MaxConnections)
	poolConfig.MinConns = int32(dbConfig.MinConnections)
	poolConfig.MaxConnLifetime = dbConfig.MaxConnLifetime
	poolConfig.MaxConnIdleTime = dbConfig.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = dbConfig.HealthCheckPeriod

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
