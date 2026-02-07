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

	router.GET("/", middleware.RequireAuth(jwtConfig), handler.Me)
	router.PUT("/", middleware.RequireAuth(jwtConfig), handler.Update)
	router.PATCH("/", middleware.RequireAuth(jwtConfig), handler.Patch)
	router.GET("/groups", middleware.RequireAuth(jwtConfig), handler.GetGroups)
	router.GET("/admin", middleware.RequireAuth(jwtConfig), handler.GetAdmin)
}
