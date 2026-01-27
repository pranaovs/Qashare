package middleware

import (
	"errors"
	"net/http"

	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			utils.LogWarn(ctx, "Authentication failed", "error", err.Error(), "path", c.Request.URL.Path)
			
			// Map specific JWT errors to appropriate error codes
			errResp := models.ErrorResponse{}
			switch {
			case errors.Is(err, utils.ErrExpiredToken):
				errResp = models.NewSimpleErrorResponse(err.Error(), models.ErrCodeExpiredToken)
			case errors.Is(err, utils.ErrAuthHeaderMissing):
				errResp = models.NewSimpleErrorResponse(err.Error(), models.ErrCodeUnauthorized)
			case errors.Is(err, utils.ErrInvalidToken), errors.Is(err, utils.ErrInvalidClaims):
				errResp = models.NewSimpleErrorResponse(err.Error(), models.ErrCodeInvalidToken)
			default:
				errResp = models.NewSimpleErrorResponse(err.Error(), models.ErrCodeInvalidToken)
			}
			
			utils.AbortWithError(c, http.StatusUnauthorized, errResp)
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

	userIDStr, ok := userID.(string)
	if !ok {
		return "", false
	}

	return userIDStr, true
}

// MustGetUserID retrieves the user ID from the context. Intended for use in handlers
// If the user ID is not found, it panics, indicating a server-side misconfiguration.
func MustGetUserID(c *gin.Context) string {
	userID, ok := GetUserID(c)
	if !ok {
		// not a runtime user error. Gin will recover and return 500.
		panic("MustGetUserID: user ID not found in context. Did you forget to add the RequireAuth middleware?")
	}
	return userID
}
