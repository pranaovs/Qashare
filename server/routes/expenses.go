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
	router.Use(middleware.RequireAuth(jwtConfig))

	router.GET("/:id", middleware.VerifyExpenseAccess(pool), handler.Get)
	router.PUT("/:id", middleware.VerifyExpenseAdmin(pool), handler.Update)
	router.PATCH("/:id", middleware.VerifyExpenseAdmin(pool), handler.Patch)
	router.DELETE("/:id", middleware.VerifyExpenseAdmin(pool), handler.Delete)
}
