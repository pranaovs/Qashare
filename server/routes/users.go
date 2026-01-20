package routes

import (
	"context"
	"errors"
	"net/http"

	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterUsersRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	// User details
	router.GET("/:id", func(c *gin.Context) {
		qUserID := c.Param("id")

		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		err = db.UsersRelated(context.Background(), pool, userID, qUserID)
		if err != nil {
			if errors.Is(err, db.ErrUsersNotRelated) {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// At this point, users are related

		result, err := db.GetUser(context.Background(), pool, qUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	// User details from email
	router.GET("/search/email/:email", func(c *gin.Context) {
		// Authenticate requester
		_, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		email, err := utils.ValidateEmail(c.Param("email"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
			return
		}
		user, err := db.GetUserFromEmail(c.Request.Context(), pool, email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	})
}
