package main

import (
	"context"
	"fmt"
	"log/slog"
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
	// Initialize pretty logger early so config-loading logs are formatted
	utils.InitDefaultLogger()

	if err := run(); err != nil {
		slog.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Re-initialize logger with config (applies debug level if set)
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
		return fmt.Errorf("invalid API_PUBLIC_URL: %w", err)
	}

	docs.SwaggerInfo.Host = u.Host
	docs.SwaggerInfo.BasePath = cfg.API.BasePath
	docs.SwaggerInfo.Schemes = []string{u.Scheme}

	// Start periodic cleanup of expired refresh tokens
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	db.StartTokenCleanup(cleanupCtx, pool, cfg.JWT.TokenCleanupFreq)

	// Setup HTTP router
	router := gin.Default()
	if err := router.SetTrustedProxies(cfg.API.TrustedProxies); err != nil {
		slog.Error("Invalid trusted proxies configuration", "error", err)
		return err
	}
	routes.RegisterRoutes(cfg.API.BasePath, router, pool, cfg.JWT, cfg.App)

	// Start server with graceful shutdown
	return startServer(router, cfg.API)
}

func initDatabase(dbConfig config.DatabaseConfig) (*pgxpool.Pool, error) {
	slog.Info("Initializing database connection...")

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
	slog.Info("Database health check passed")

	// Run migrations
	if err := db.Migrate(pool, dbConfig.MigrationsDir); err != nil {
		db.Close(pool)
		return nil, err
	}

	// Verify migration integrity (optional, can be disabled via env var)
	if dbConfig.VerifyMigrations {
		verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer verifyCancel()
		if err := db.VerifyMigrationIntegrity(verifyCtx, pool, dbConfig.MigrationsDir); err != nil {
			slog.Warn("Migration integrity check failed", "error", err)
			// Non-fatal warning - allow startup but log the issue
		}
	}

	slog.Info("Database initialized successfully")
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
		slog.Info("Server starting", "port", apiConfig.BindPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("Server stopped gracefully")
	return nil
}
