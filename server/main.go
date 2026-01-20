package main

import (
	"log"

	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/routes"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	utils.Loadenv()

	// Open database
	pool, err := db.Connect(utils.Getenv("DB_URL", "postgres://postgres:postgres@localhost:5432/shared_expenses"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Run Migrations
	if err := db.Migrate(pool, utils.Getenv("DB_MIGRATIONS_DIR", "db/migrations")); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	routes.RegisterRoutes(router, pool)

	port := utils.Getenv("API_PORT", "8080")
	log.Println("Server running on port", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
