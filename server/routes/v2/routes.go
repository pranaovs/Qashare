package v2

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, appConfig config.AppConfig, jwtConfig config.JWTConfig) {
	groupsHandler := NewGroupsHandler(pool, appConfig)

	// Groups — paginated list endpoints
	groups := router.Group("/groups")
	groups.Use(middleware.RequireAuth(jwtConfig))
	groups.GET("/:id/expenses", middleware.RequireGroupMember(pool), groupsHandler.GetExpenses)
	groups.GET("/:id/settlements", middleware.RequireGroupMember(pool), groupsHandler.GetSettlements)
	groups.GET("/:id/spendings", middleware.RequireGroupMember(pool), groupsHandler.GetSpendings)
}
