package routes

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/handlers"
	"github.com/pranaovs/qashare/routes/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterSettlementsRoutes(router *gin.RouterGroup, pool *pgxpool.Pool, jwtConfig config.JWTConfig, appConfig config.AppConfig) {
	handler := handlers.NewSettlementsHandler(pool, appConfig)
	router.Use(middleware.RequireAuth(jwtConfig))

	router.POST("/", handler.Create)
	router.GET("/:id", middleware.VerifySettlementAccess(pool), handler.Get)
	router.PUT("/:id", middleware.VerifySettlementAdmin(pool), handler.Update)
	router.PATCH("/:id", middleware.VerifySettlementAdmin(pool), handler.Patch)
	router.DELETE("/:id", middleware.VerifySettlementAdmin(pool), handler.Delete)
}
