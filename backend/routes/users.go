package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterUserRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	router.GET("list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "list of users"})
	})
}
