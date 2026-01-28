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
	"github.com/pranaovs/qashare/docs"
	"github.com/pranaovs/qashare/routes"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @title Qashare API
// @version 1.0
// @description API for managing shared expenses, groups, and user authentication

// @contact.name Pranaov S
// @contact.email qashare.contact@pranaovs.me

// @license.name AGPL-3.0
// @license.url https://www.gnu.org/licenses/agpl-3.0.en.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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

	// Swagger url setup
	docs.SwaggerInfo.Host = utils.GetEnv("API_HOST", "localhost") + ":" + strconv.Itoa(utils.GetEnvPort("API_PORT", 8080))
	docs.SwaggerInfo.BasePath = utils.GetEnv("API_BASE_PATH", "/api")

	// Setup HTTP router
	router := setupRouter(pool)

	// Start server with graceful shutdown
	return startServer(router)
}

func initDatabase() (*pgxpool.Pool, error) {
	log.Println("[INIT] Initializing database connection...")

	// Get database configuration from environment
	dbURL := utils.GetEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/qashare")

	// Connects to the PostgreSQL database using the provided URL. The database must already exist.
	pool, err := db.Connect(dbURL)
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
	log.Println("[INIT] Database health check passed")

	// Run migrations
	migrationsDir := utils.GetEnv("DB_MIGRATIONS_DIR", "migrations")
	if err := db.Migrate(pool, migrationsDir); err != nil {
		db.Close(pool)
		return nil, err
	}

	// Verify migration integrity (optional, can be disabled via env var)
	if utils.GetEnvBool("DB_VERIFY_MIGRATIONS", true) {
		if err := db.VerifyMigrationIntegrity(ctx, pool, migrationsDir); err != nil {
			log.Printf("[INIT] Migration integrity check failed: %v", err)
			// Non-fatal warning - allow startup but log the issue
		}
	}

	log.Println("[INIT] Database initialized successfully")
	return pool, nil
}

func setupRouter(pool *pgxpool.Pool) *gin.Engine {
	router := gin.Default()
	routes.RegisterRoutes(router, pool)
	return router
}

func startServer(router *gin.Engine) error {
	port := utils.GetEnvPort("API_PORT", 8080)
	srv := &http.Server{
		Addr:    utils.GetEnv("API_HOST", "localhost") + ":" + strconv.Itoa(port),
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %d", port)
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
