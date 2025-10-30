package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(pool *pgxpool.Pool, migrationsDir string) error {
	ctx := context.Background()

	// Create migrations tracking table if it doesn't exist
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			migration_name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	files := []string{}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(files)

	// Apply migrations
	appliedCount := 0
	for _, file := range files {
		migrationName := filepath.Base(file)

		// Check if migration was already applied
		var exists bool
		err = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE migration_name = $1)`,
			migrationName,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration status for %s: %w", migrationName, err)
		}

		if exists {
			log.Printf("[MIGRATIONS] Skipping already applied: %s", migrationName)
			continue
		}

		// Apply migration in a transaction
		log.Printf("[MIGRATIONS] Applying: %s", migrationName)
		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read file %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", migrationName, err)
		}

		_, err = tx.Exec(ctx, string(sql))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("execute %s: %w", migrationName, err)
		}

		// Record migration as applied
		_, err = tx.Exec(ctx,
			`INSERT INTO schema_migrations (migration_name) VALUES ($1)`,
			migrationName,
		)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", migrationName, err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("commit transaction for %s: %w", migrationName, err)
		}

		appliedCount++
	}

	if appliedCount > 0 {
		log.Printf("[MIGRATIONS] Successfully applied %d new migration(s)", appliedCount)
	} else {
		log.Println("[MIGRATIONS] No new migrations to apply")
	}

	return nil
}
