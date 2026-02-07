package routes

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterExpensesRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, jwtConfig config.JWTConfig, appConfig config.AppConfig) {
	handler := handlers.NewExpensesHandler(pool, appConfig)

	router.POST("/", middleware.RequireAuth(jwtConfig), handler.Create)
	router.GET("/:id", middleware.RequireAuth(jwtConfig), middleware.VerifyExpenseAccess(pool), handler.GetExpense)
	router.PUT("/:id", middleware.RequireAuth(jwtConfig), middleware.VerifyExpenseAdmin(pool), handler.Update)
	router.PATCH("/:id", middleware.RequireAuth(jwtConfig), middleware.VerifyExpenseAdmin(pool), handler.Patch)
	router.DELETE("/:id", middleware.RequireAuth(jwtConfig), middleware.VerifyExpenseAdmin(pool), handler.Delete)
}
