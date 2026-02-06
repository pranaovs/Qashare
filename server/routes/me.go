package routes

import (
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterMeRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	handler := handlers.NewMeHandler(pool)

	router.GET("/", middleware.RequireAuth(), handler.Me)
	router.GET("/groups", middleware.RequireAuth(), handler.ListGroups)
	router.GET("/admin", middleware.RequireAuth(), handler.ListAdmin)
}
