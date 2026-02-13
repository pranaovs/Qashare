package middleware

import (
	"github.com/google/uuid"
	"github.com/pranaovs/qashare/apperrors"
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

func RequireAuth(jwtConfig config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"), jwtConfig)
		if err != nil {
			utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
				utils.ErrInvalidToken: apierrors.ErrInvalidToken,
			}))
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.UUID{}, false
	}

	userIDVal, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.UUID{}, false
	}

	return userIDVal, true
}

// MustGetUserID retrieves the user ID from the context. Intended for use in handlers
// If the user ID is not found, it panics, indicating a server-side misconfiguration.
func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, ok := GetUserID(c)
	if !ok {
		// not a runtime user error. Gin will recover and return 500.
		panic("MustGetUserID: user ID not found in context. Did you forget to add the RequireAuth middleware?")
	}
	return userID
}
