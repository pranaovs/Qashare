package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxFunc is a function that executes within a database transaction
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// WithTransaction executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
// This provides a consistent pattern for transaction management.
func WithTransaction(ctx context.Context, pool *pgxpool.Pool, fn TxFunc) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is properly handled
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			_ = tx.Rollback(ctx)
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			// Rollback on error
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Printf("[DB] Failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	// Execute the function
	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecuteInBatch executes multiple queries in a batch for better performance.
// Useful for bulk inserts, updates, or deletes.
func ExecuteInBatch(ctx context.Context, pool *pgxpool.Pool, queries []BatchQuery) error {
	if len(queries) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, q := range queries {
		batch.Queue(q.SQL, q.Args...)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	// Execute all queries and collect errors
	for i := 0; i < len(queries); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch query %d failed: %w", i, err)
		}
	}

	return nil
}

// BatchQuery represents a single query in a batch operation
type BatchQuery struct {
	SQL  string
	Args []interface{}
}

// RecordExists checks if a record exists in a table with the given condition
// Example: exists, err := RecordExists(ctx, pool, "users", "email = $1", email)
func RecordExists(ctx context.Context, pool *pgxpool.Pool, table, condition string, args ...interface{}) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)", table, condition)

	var exists bool
	err := pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if record exists: %w", err)
	}

	return exists, nil
}

// CountRecords returns the count of records in a table matching the condition
// Example: count, err := CountRecords(ctx, pool, "users", "is_guest = $1", true)
func CountRecords(ctx context.Context, pool *pgxpool.Pool, table, condition string, args ...interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, condition)

	var count int64
	err := pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// LogQuery logs a database query with its parameters (for debugging)
// This should only be used in development, not in production
func LogQuery(query string, args ...interface{}) {
	log.Printf("[DB QUERY] %s [args: %v]", query, args)
}

// MeasureQueryTime measures and logs the execution time of a query
// Returns a function that should be deferred to measure the time
func MeasureQueryTime(operation string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		log.Printf("[DB TIMING] %s took %v", operation, duration)
	}
}

// RetryOnError retries a database operation if it fails with a transient error
// Useful for handling temporary connection issues
func RetryOnError(ctx context.Context, maxRetries int, operation func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		// Check if error is retryable (connection errors, timeouts, etc.)
		if !isRetryableError(err) {
			return err
		}

		// Wait before retrying with exponential backoff
		if i < maxRetries-1 {
			waitTime := time.Duration(1<<uint(i)) * 100 * time.Millisecond
			log.Printf("[DB] Operation failed, retrying in %v (attempt %d/%d): %v",
				waitTime, i+1, maxRetries, err)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

// isRetryableError checks if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Check for common transient errors
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"temporary",
		"too many connections",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// ValidateUUID checks if a string is a valid UUID format
// This is a basic validation that checks format structure
// For more rigorous validation, consider using github.com/google/uuid package
func ValidateUUID(uuid string) bool {
	if len(uuid) != 36 {
		return false
	}
	// Basic format check: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		return false
	}

	// Check that all other characters are valid hexadecimal
	for i, c := range uuid {
		// Skip the dash positions
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		// Check if character is a valid hex digit
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}
