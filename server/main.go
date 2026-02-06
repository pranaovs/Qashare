package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/pranaovs/qashare/config"
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
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Initialize logger
	utils.InitLogger(cfg)

	// Initialize database with enhanced configuration
	pool, err := initDatabase(cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close(pool)

	// Swagger url setup
	u, err := url.Parse(cfg.API.PublicURL)
	if err != nil {
		log.Fatalf("Invalid API_PUBLIC_URL: %v", err)
	}

	docs.SwaggerInfo.Host = u.Host
	docs.SwaggerInfo.BasePath = cfg.API.BasePath
	docs.SwaggerInfo.Schemes = []string{u.Scheme}

	// Setup HTTP router
	router := gin.Default()
	routes.RegisterRoutes(cfg.API.BasePath, router, pool, cfg.JWT, cfg.App)

	// Start server with graceful shutdown
	return startServer(router, cfg.API)
}

func initDatabase(dbConfig config.DatabaseConfig) (*pgxpool.Pool, error) {
	log.Println("[INIT] Initializing database connection...")

	// Connects to the PostgreSQL database using the provided URL. The database must already exist.
	pool, err := db.Connect(dbConfig)
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
	if err := db.Migrate(pool, dbConfig.MigrationsDir); err != nil {
		db.Close(pool)
		return nil, err
	}

	// Verify migration integrity (optional, can be disabled via env var)
	if dbConfig.VerifyMigrations {
		if err := db.VerifyMigrationIntegrity(ctx, pool, dbConfig.MigrationsDir); err != nil {
			log.Printf("[INIT] Migration integrity check failed: %v", err)
			// Non-fatal warning - allow startup but log the issue
		}
	}

	log.Println("[INIT] Database initialized successfully")
	return pool, nil
}

func startServer(router *gin.Engine, apiConfig config.APIConfig) error {
	srv := &http.Server{
		Addr:    apiConfig.BindAddr + ":" + strconv.Itoa(apiConfig.BindPort),
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %d", apiConfig.BindPort)
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
