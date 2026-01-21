package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/routes"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load environment variables
	utils.Loadenv()

	// Initialize database with enhanced configuration
	pool, err := initDatabase()
	if err != nil {
		return err
	}
	defer db.Close(pool)

	// Setup HTTP router
	router := setupRouter(pool)

	// Start server with graceful shutdown
	return startServer(router)
}

func initDatabase() (*pgxpool.Pool, error) {
	log.Println("[INIT] Initializing database connection...")

	// Get database configuration from environment
	dbURL := utils.Getenv("DB_URL", "postgres://postgres:postgres@localhost:5432/shared_expenses")

	// Create database config with optional environment overrides
	config := createDBConfig(dbURL)

	// Connect to database (will auto-create if not exists)
	pool, err := db.ConnectWithConfig(config)
	if err != nil {
		return nil, err
	}

	// Perform health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.HealthCheck(ctx, pool); err != nil {
		db.Close(pool)
		return nil, err
	}
	log.Println("[INIT] ✓ Database health check passed")

	// Run migrations
	migrationsDir := utils.Getenv("DB_MIGRATIONS_DIR", "db/migrations")
	if err := db.Migrate(pool, migrationsDir); err != nil {
		db.Close(pool)
		return nil, err
	}

	// Verify migration integrity (optional, can be disabled via env var)
	if utils.Getenv("DB_VERIFY_MIGRATIONS", "true") == "true" {
		if err := db.VerifyMigrationIntegrity(ctx, pool, migrationsDir); err != nil {
			log.Printf("[INIT] ⚠ Migration integrity check failed: %v", err)
			// Non-fatal warning - allow startup but log the issue
		}
	}

	log.Println("[INIT] ✓ Database initialized successfully")
	return pool, nil
}

// createDBConfig creates a database configuration with optional environment overrides
func createDBConfig(dbURL string) *db.DBConfig {
	config := db.DefaultDBConfig(dbURL)

	// Allow environment variable overrides for connection pool settings
	if maxConn := utils.Getenv("DB_MAX_CONNECTIONS", ""); maxConn != "" {
		if val, err := strconv.Atoi(maxConn); err == nil && val > 0 {
			config.MaxConnections = int32(val)
		} else {
			log.Printf("[INIT] ⚠ Invalid DB_MAX_CONNECTIONS value '%s', using default: %d", maxConn, config.MaxConnections)
		}
	}

	if minConn := utils.Getenv("DB_MIN_CONNECTIONS", ""); minConn != "" {
		if val, err := strconv.Atoi(minConn); err == nil && val > 0 {
			config.MinConnections = int32(val)
		} else {
			log.Printf("[INIT] ⚠ Invalid DB_MIN_CONNECTIONS value '%s', using default: %d", minConn, config.MinConnections)
		}
	}

	if timeout := utils.Getenv("DB_CONNECT_TIMEOUT", ""); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil && val > 0 {
			config.ConnectTimeout = time.Duration(val) * time.Second
		} else {
			log.Printf("[INIT] ⚠ Invalid DB_CONNECT_TIMEOUT value '%s', using default: %v", timeout, config.ConnectTimeout)
		}
	}

	return config
}

func setupRouter(pool *pgxpool.Pool) *gin.Engine {
	router := gin.Default()
	routes.RegisterRoutes(router, pool)
	return router
}

func startServer(router *gin.Engine) error {
	port := utils.Getenv("API_PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server stopped gracefully")
	return nil
}
