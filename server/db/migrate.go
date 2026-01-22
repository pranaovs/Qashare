package db

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationInfo holds metadata about a database migration
type MigrationInfo struct {
	Name      string
	AppliedAt time.Time
	Checksum  string
}

// MigrationStatus represents the current state of migrations
type MigrationStatus struct {
	TotalMigrations   int
	AppliedMigrations int
	PendingMigrations int
	Migrations        []MigrationInfo
}

// Migrate applies all pending database migrations from the specified directory.
// It tracks applied migrations in the schema_migrations table and ensures idempotent execution.
// Migrations are applied in alphabetical order by filename.
func Migrate(pool *pgxpool.Pool, migrationsDir string) error {
	ctx := context.Background()

	log.Printf("[MIGRATIONS] Starting migration process from directory: %s", migrationsDir)

	// Initialize migration tracking table
	if err := initMigrationTable(ctx, pool); err != nil {
		return err
	}

	// Get list of migration files
	migrationFiles, err := getMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	if len(migrationFiles) == 0 {
		log.Println("[MIGRATIONS] No migration files found")
		return nil
	}

	log.Printf("[MIGRATIONS] Found %d migration file(s)", len(migrationFiles))

	// Apply each migration
	appliedCount := 0
	for _, file := range migrationFiles {
		applied, err := applyMigration(ctx, pool, file)
		if err != nil {
			return err
		}
		if applied {
			appliedCount++
		}
	}

	// Log summary
	if appliedCount > 0 {
		log.Printf("[MIGRATIONS] Successfully applied %d new migration(s)", appliedCount)
	} else {
		log.Println("[MIGRATIONS] Database is up to date - no new migrations to apply")
	}

	return nil
}

// initMigrationTable creates the schema_migrations table if it doesn't exist
func initMigrationTable(ctx context.Context, pool *pgxpool.Pool) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			migration_name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			checksum TEXT NOT NULL,
			execution_time_ms INTEGER
		)
	`

	_, err := pool.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	log.Println("[MIGRATIONS] Migration tracking table initialized")
	return nil
}

// getMigrationFiles reads and returns a sorted list of migration files from the directory
func getMigrationFiles(migrationsDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory '%s': %w", migrationsDir, err)
	}

	files := []string{}
	for _, e := range entries {
		// Only include .sql files
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}

	// Sort files alphabetically to ensure consistent ordering
	sort.Strings(files)
	return files, nil
}

// applyMigration applies a single migration file if it hasn't been applied yet
// Returns true if migration was applied, false if it was skipped
func applyMigration(ctx context.Context, pool *pgxpool.Pool, filePath string) (bool, error) {
	migrationName := filepath.Base(filePath)

	// Check if migration was already applied
	alreadyApplied, err := isMigrationApplied(ctx, pool, migrationName)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status for '%s': %w", migrationName, err)
	}

	if alreadyApplied {
		log.Printf("[MIGRATIONS] ⊘ Skipping already applied: %s", migrationName)
		return false, nil
	}

	// Read migration file content
	sqlContent, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read migration file '%s': %w", filePath, err)
	}

	// Calculate checksum for integrity verification
	checksum := calculateChecksum(sqlContent)

	// Execute migration in a transaction
	log.Printf("[MIGRATIONS] ⟳ Applying migration: %s", migrationName)
	startTime := time.Now()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction for '%s': %w", migrationName, err)
	}

	// Ensure transaction is rolled back on error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Printf("[MIGRATIONS] ✗ Failed to rollback transaction for '%s': %v", migrationName, rbErr)
			}
		}
	}()

	// Execute the migration SQL
	_, err = tx.Exec(ctx, string(sqlContent))
	if err != nil {
		return false, fmt.Errorf("failed to execute migration '%s': %w", migrationName, err)
	}

	// Calculate execution time
	executionTime := time.Since(startTime).Milliseconds()

	// Record migration as applied
	_, err = tx.Exec(ctx,
		`INSERT INTO schema_migrations (migration_name, applied_at, checksum, execution_time_ms)
		 VALUES ($1, $2, $3, $4)`,
		migrationName,
		time.Now(),
		checksum,
		executionTime,
	)
	if err != nil {
		return false, fmt.Errorf("failed to record migration '%s': %w", migrationName, err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit transaction for '%s': %w", migrationName, err)
	}

	log.Printf("[MIGRATIONS] Successfully applied: %s (took %dms)", migrationName, executionTime)
	return true, nil
}

// isMigrationApplied checks if a migration has already been applied
func isMigrationApplied(ctx context.Context, pool *pgxpool.Pool, migrationName string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE migration_name = $1)`,
		migrationName,
	).Scan(&exists)

	return exists, err
}

// calculateChecksum computes SHA-256 checksum of the migration content
func calculateChecksum(content []byte) string {
	hash := sha256.New()
	hash.Write(content)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// GetMigrationStatus returns the current status of all migrations
func GetMigrationStatus(ctx context.Context, pool *pgxpool.Pool) (*MigrationStatus, error) {
	rows, err := pool.Query(ctx,
		`SELECT migration_name, applied_at, checksum 
		 FROM schema_migrations 
		 ORDER BY applied_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query migration status: %w", err)
	}
	defer rows.Close()

	status := &MigrationStatus{
		Migrations: make([]MigrationInfo, 0),
	}

	for rows.Next() {
		var info MigrationInfo
		if err := rows.Scan(&info.Name, &info.AppliedAt, &info.Checksum); err != nil {
			return nil, fmt.Errorf("failed to scan migration info: %w", err)
		}
		status.Migrations = append(status.Migrations, info)
	}

	status.AppliedMigrations = len(status.Migrations)
	return status, nil
}

// VerifyMigrationIntegrity checks if applied migrations match their recorded checksums
func VerifyMigrationIntegrity(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	log.Println("[MIGRATIONS] Verifying migration integrity...")

	status, err := GetMigrationStatus(ctx, pool)
	if err != nil {
		return err
	}

	for _, migration := range status.Migrations {
		filePath := filepath.Join(migrationsDir, migration.Name)

		// Read current file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("[MIGRATIONS] Warning: Migration file '%s' no longer exists", migration.Name)
				continue
			}
			return fmt.Errorf("failed to read migration file '%s': %w", migration.Name, err)
		}

		// Calculate and compare checksum
		currentChecksum := calculateChecksum(content)
		if currentChecksum != migration.Checksum {
			return fmt.Errorf("integrity check failed for '%s': checksum mismatch (expected: %s, got: %s)",
				migration.Name, migration.Checksum, currentChecksum)
		}
	}

	log.Printf("[MIGRATIONS] Integrity verification passed for %d migration(s)", len(status.Migrations))
	return nil
}
