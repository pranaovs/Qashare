package routes

import (
	"github.com/pranaovs/qashare/handlers"
	"github.com/pranaovs/qashare/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterGroupsRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	handler := handlers.NewGroupsHandler(pool)

	router.POST("/", middleware.RequireAuth(), handler.Create)
	router.GET("/me", middleware.RequireAuth(), handler.ListUserGroups)
	router.GET("/admin", middleware.RequireAuth(), handler.ListAdminGroups)
	router.GET("/:id", middleware.RequireAuth(), handler.GetGroup)
	router.POST("/:id/members", middleware.RequireAuth(), handler.AddMembers)
	router.DELETE("/:id/members", middleware.RequireAuth(), handler.RemoveMembers)
}
