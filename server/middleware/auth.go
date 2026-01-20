package middleware

import (
	"net/http"

	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	
	id, ok := userID.(string)
	return id, ok
}
