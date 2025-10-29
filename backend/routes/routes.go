package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(router *gin.Engine, pool *pgxpool.Pool) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	RegisterUsersRoutes(router.Group("/users"), pool)
}
