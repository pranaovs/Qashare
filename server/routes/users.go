package routes

import (
	"github.com/pranaovs/qashare/handlers"
	"github.com/pranaovs/qashare/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterUsersRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	handler := handlers.NewUsersHandler(pool)

	router.GET("/:id", middleware.RequireAuth(), handler.GetUser)
	router.GET("/search/email/:email", middleware.RequireAuth(), handler.SearchByEmail)
}
