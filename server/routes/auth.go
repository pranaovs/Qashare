package routes

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterAuthRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, appConfig config.AppConfig, jwtConfig config.JWTConfig) {
	handler := handlers.NewAuthHandler(pool, appConfig, jwtConfig)

	router.POST("/register", handler.Register)
	router.GET("/verify", handler.Verify)
	router.POST("/login", handler.Login)
	router.POST("/refresh", handler.Refresh)
	router.POST("/logout", middleware.RequireAuth(jwtConfig), handler.Logout)
	router.POST("/logout-all", middleware.RequireAuth(jwtConfig), handler.LogoutAll)
}
