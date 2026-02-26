package middleware

import (
	"github.com/google/uuid"
	"github.com/pranaovs/qashare/apperrors"
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey    = "userID"
	SessionIDKey = "sessionID"
)

func RequireAuth(jwtConfig config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := utils.ExtractAccessClaims(c.GetHeader("Authorization"), jwtConfig)
		if err != nil {
			utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
				utils.ErrExpiredToken: apierrors.ErrExpiredToken,
				utils.ErrInvalidToken: apierrors.ErrInvalidToken,
			}))
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			utils.SendError(c, apierrors.ErrInvalidToken)
			c.Abort()
			return
		}

		sessionID, err := uuid.Parse(claims.SessionID)
		if err != nil {
			utils.SendError(c, apierrors.ErrInvalidToken)
			c.Abort()
			return
		}

		c.Set(UserIDKey, userID)
		c.Set(SessionIDKey, sessionID)
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

// MustGetUserID retrieves the user ID from the context. Intended for use in handlers.
// If the user ID is not found, it panics, indicating a server-side misconfiguration.
func MustGetUserID(c *gin.Context) uuid.UUID {
	userID, ok := GetUserID(c)
	if !ok {
		panic("MustGetUserID: user ID not found in context. Did you forget to add the RequireAuth middleware?")
	}
	return userID
}

func GetSessionID(c *gin.Context) (uuid.UUID, bool) {
	sessionID, exists := c.Get(SessionIDKey)
	if !exists {
		return uuid.UUID{}, false
	}

	sessionIDVal, ok := sessionID.(uuid.UUID)
	if !ok {
		return uuid.UUID{}, false
	}

	return sessionIDVal, true
}

// MustGetSessionID retrieves the session ID from the context. Intended for use in handlers.
// If the session ID is not found, it panics, indicating a server-side misconfiguration.
func MustGetSessionID(c *gin.Context) uuid.UUID {
	sessionID, ok := GetSessionID(c)
	if !ok {
		panic("MustGetSessionID: session ID not found in context. Did you forget to add the RequireAuth middleware?")
	}
	return sessionID
}
