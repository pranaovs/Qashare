package routes

import (
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterAuthRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	handler := handlers.NewAuthHandler(pool)

	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	router.GET("/me", middleware.RequireAuth(), handler.Me)
	router.POST("/guest", middleware.RequireAuth(), handler.RegisterGuest)
}
