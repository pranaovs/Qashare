package routes

import (
	"github.com/pranaovs/qashare/handlers"
	"github.com/pranaovs/qashare/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterAuthRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	handler := handlers.NewAuthHandler(pool)

	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	router.GET("/me", middleware.RequireAuth(), handler.Me)
}
