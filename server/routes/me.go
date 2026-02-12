package routes

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterMeRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, jwtConfig config.JWTConfig) {
	handler := handlers.NewMeHandler(pool)
	router.Use(middleware.RequireAuth(jwtConfig))

	router.GET("/", handler.Me)
	router.PUT("/", handler.Update)
	router.PATCH("/", handler.Patch)
	router.DELETE("/", handler.Delete)
	router.GET("/groups", handler.GetGroups)
	router.GET("/admin", handler.GetOwner)
}
