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

	router.POST("/", middleware.RequireAuth(jwtConfig), handler.Create)
	router.GET("/me", middleware.RequireAuth(jwtConfig), handler.ListUser)
	router.GET("/admin", middleware.RequireAuth(jwtConfig), handler.ListAdmin)
	router.GET("/:id", middleware.RequireAuth(jwtConfig), middleware.RequireGroupMember(pool), handler.Get)
	router.PUT("/:id", middleware.RequireAuth(jwtConfig), middleware.RequireGroupAdmin(pool), handler.Update)
	router.PATCH("/:id", middleware.RequireAuth(jwtConfig), middleware.RequireGroupAdmin(pool), handler.Patch)
	router.POST("/:id/members", middleware.RequireAuth(jwtConfig), middleware.RequireGroupAdmin(pool), handler.AddMembers)
	router.DELETE("/:id/members", middleware.RequireAuth(jwtConfig), middleware.RequireGroupAdmin(pool), handler.RemoveMembers)
	router.GET("/:id/expenses", middleware.RequireAuth(jwtConfig), middleware.RequireGroupMember(pool), handler.GetExpenses)
	router.GET("/:id/settlements", middleware.RequireAuth(jwtConfig), middleware.RequireGroupMember(pool), handler.GetSettlements)
	router.GET("/:id/spendings", middleware.RequireAuth(jwtConfig), middleware.RequireGroupMember(pool), handler.GetSpendings)
}
