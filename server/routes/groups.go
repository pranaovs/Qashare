package routes

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterGroupsRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, jwtConfig config.JWTConfig, appConfig config.AppConfig) {
	handler := handlers.NewGroupsHandler(pool, appConfig)
	router.Use(middleware.RequireAuth(jwtConfig))

	router.POST("/", handler.Create)
	router.GET("/me", handler.ListUser)
	router.GET("/admin", handler.ListAdmin)
	router.GET("/:id", middleware.RequireGroupMember(pool), handler.Get)
	router.PUT("/:id", middleware.RequireGroupAdmin(pool), handler.Update)
	router.PATCH("/:id", middleware.RequireGroupAdmin(pool), handler.Patch)
	router.DELETE("/:id", middleware.RequireGroupAdmin(pool), handler.Delete)
	router.POST("/:id/members", middleware.RequireGroupAdmin(pool), handler.AddMembers)
	router.DELETE("/:id/members", middleware.RequireGroupAdmin(pool), handler.RemoveMembers)
	router.GET("/:id/expenses", middleware.RequireGroupMember(pool), handler.GetExpenses)
	router.GET("/:id/settlements", middleware.RequireGroupMember(pool), handler.GetSettlements)
	router.GET("/:id/spendings", middleware.RequireGroupMember(pool), handler.GetSpendings)
}
