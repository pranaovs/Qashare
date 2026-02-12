package routes

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterUsersRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, jwtConfig config.JWTConfig) {
	handler := handlers.NewUsersHandler(pool)
	router.Use(middleware.RequireAuth(jwtConfig))

	router.GET("/:id", handler.Get)
	router.GET("/search/email/:email", handler.SearchByEmail)
	router.POST("/guest", handler.RegisterGuest)
}
