package v1

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, appConfig config.AppConfig, jwtConfig config.JWTConfig) {
	authHandler := NewAuthHandler(pool, appConfig, jwtConfig)
	meHandler := NewMeHandler(pool, appConfig)
	usersHandler := NewUsersHandler(pool, appConfig)
	groupsHandler := NewGroupsHandler(pool, appConfig)
	expensesHandler := NewExpensesHandler(pool, appConfig)
	settlementsHandler := NewSettlementsHandler(pool, appConfig)

	// Auth (no auth middleware on most routes)
	auth := router.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.GET("/verify", authHandler.Verify)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.POST("/logout", middleware.RequireAuth(jwtConfig), authHandler.Logout)
	auth.POST("/logout-all", middleware.RequireAuth(jwtConfig), authHandler.LogoutAll)

	// Me
	me := router.Group("/me")
	me.Use(middleware.RequireAuth(jwtConfig))
	me.GET("/", meHandler.Me)
	me.PUT("/", meHandler.Update)
	me.PATCH("/", meHandler.Patch)
	me.DELETE("/", meHandler.Delete)
	me.GET("/groups", meHandler.GetGroups)
	me.GET("/admin", meHandler.GetOwner)

	// Users
	users := router.Group("/users")
	users.Use(middleware.RequireAuth(jwtConfig))
	users.GET("/:id", usersHandler.Get)
	users.GET("/search/email/:email", usersHandler.SearchByEmail)
	users.POST("/guest", usersHandler.RegisterGuest)

	// Groups
	groups := router.Group("/groups")
	groups.Use(middleware.RequireAuth(jwtConfig))
	groups.POST("/", groupsHandler.Create)
	groups.GET("/:id", middleware.RequireGroupMember(pool), groupsHandler.Get)
	groups.PUT("/:id", middleware.RequireGroupAdmin(pool), groupsHandler.Update)
	groups.PATCH("/:id", middleware.RequireGroupAdmin(pool), groupsHandler.Patch)
	groups.DELETE("/:id", middleware.RequireGroupAdmin(pool), groupsHandler.Delete)
	groups.POST("/:id/members", middleware.RequireGroupAdmin(pool), groupsHandler.AddMembers)
	groups.DELETE("/:id/members", middleware.RequireGroupAdmin(pool), groupsHandler.RemoveMembers)
	groups.GET("/:id/expenses", middleware.RequireGroupMember(pool), groupsHandler.GetExpenses)
	groups.POST("/:id/expenses", middleware.RequireGroupMember(pool), expensesHandler.Create)
	groups.GET("/:id/settle", middleware.RequireGroupMember(pool), groupsHandler.GetSettle)
	groups.POST("/:id/settle", middleware.RequireGroupMember(pool), settlementsHandler.Create)
	groups.GET("/:id/settlements", middleware.RequireGroupMember(pool), groupsHandler.GetSettlements)
	groups.GET("/:id/spendings", middleware.RequireGroupMember(pool), groupsHandler.GetSpendings)

	// Expenses (individual)
	expenses := router.Group("/expenses")
	expenses.Use(middleware.RequireAuth(jwtConfig))
	expenses.GET("/:id", middleware.VerifyExpenseAccess(pool), expensesHandler.Get)
	expenses.PUT("/:id", middleware.VerifyExpenseAdmin(pool), expensesHandler.Update)
	expenses.PATCH("/:id", middleware.VerifyExpenseAdmin(pool), expensesHandler.Patch)
	expenses.DELETE("/:id", middleware.VerifyExpenseDeleteAccess(pool), expensesHandler.Delete)

	// Settlements (individual)
	settlements := router.Group("/settlements")
	settlements.Use(middleware.RequireAuth(jwtConfig))
	settlements.GET("/:id", middleware.VerifySettlementAccess(pool), settlementsHandler.Get)
	settlements.PUT("/:id", middleware.VerifySettlementAdmin(pool), settlementsHandler.Update)
	settlements.PATCH("/:id", middleware.VerifySettlementAdmin(pool), settlementsHandler.Patch)
	settlements.DELETE("/:id", middleware.VerifySettlementAdmin(pool), settlementsHandler.Delete)
}
