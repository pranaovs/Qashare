package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/config"
	v1 "github.com/pranaovs/qashare/routes/v1"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(basepath string, router *gin.Engine, pool *pgxpool.Pool, jwtConfig config.JWTConfig, appConfig config.AppConfig) {
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true
	router.RemoveExtraSlash = true

	// Health check
	router.GET(basepath+"/health", HealthCheck)

	// Swagger documentation
	if !appConfig.DisableSwagger {
		router.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		})
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// v1 routes
	v1.RegisterRoutes(router.Group(basepath+"/v1"), pool, appConfig, jwtConfig)
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
