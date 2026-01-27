package routes

import (
	"github.com/pranaovs/qashare/handlers"
	"github.com/pranaovs/qashare/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterExpensesRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	handler := handlers.NewExpensesHandler(pool)

	router.POST("/", middleware.RequireAuth(), handler.Create)
	router.GET("/:id", middleware.RequireAuth(), middleware.VerifyExpenseAccess(pool), handler.GetExpense)
	router.PUT("/:id", middleware.RequireAuth(), middleware.VerifyExpenseAdmin(pool), handler.Update)
	router.DELETE("/:id", middleware.RequireAuth(), middleware.VerifyExpenseAdmin(pool), handler.Delete)
}
