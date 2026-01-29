package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/utils"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(basepath string, router *gin.Engine, pool *pgxpool.Pool) {
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true
	router.RemoveExtraSlash = true

	// Health check
	router.GET("/health", HealthCheck)

	// Swagger documentation
	if !utils.GetEnvBool("DISABLE_SWAGGER", false) {
		router.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		})
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Create a base route group
	baseGroup := router.Group(basepath + "/v1")

	RegisterAuthRoutes(baseGroup.Group("/auth"), pool)
	RegisterUsersRoutes(baseGroup.Group("/users"), pool)
	RegisterGroupsRoutes(baseGroup.Group("/groups"), pool)
	RegisterExpensesRoutes(baseGroup.Group("/expenses"), pool)
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API is running
// @Tags health
// @Produce plain
// @Success 200 {string} string "ok"
// @Router /v1/health [get]
func HealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
