package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/pranaovs/qashare/docs" // Import swagger docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, pool *pgxpool.Pool) {
	// Health check
	router.GET("/health", HealthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	RegisterAuthRoutes(router.Group("/auth"), pool)
	RegisterUsersRoutes(router.Group("/users"), pool)
	RegisterGroupsRoutes(router.Group("/groups"), pool)
	RegisterExpensesRoutes(router.Group("/expenses"), pool)
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API is running
// @Tags health
// @Produce plain
// @Success 200 {string} string "ok"
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
