package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	utils.Loadenv()

	pool, err := initDatabase()
	if err != nil {
		return err
	}
	defer pool.Close()

	router := setupRouter(pool)

	return startServer(router)
}

func initDatabase() (*pgxpool.Pool, error) {
	dbURL := utils.Getenv("DB_URL", "postgres://postgres:postgres@localhost:5432/shared_expenses")
	pool, err := db.Connect(dbURL)
	if err != nil {
		return nil, err
	}

	migrationsDir := utils.Getenv("DB_MIGRATIONS_DIR", "db/migrations")
	if err := db.Migrate(pool, migrationsDir); err != nil {
		pool.Close()
		return nil, err
	}

	log.Println("Database connected and migrations applied")
	return pool, nil
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
