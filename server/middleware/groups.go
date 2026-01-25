package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/db"
)

const GroupIDKey = "groupID"

// RequireGroupMember checks if the authenticated user is a member of the group
func RequireGroupMember(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)
		groupID, ok := c.Params.Get("id")

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID not provided"})
			c.Abort()
			return
		}

		ok, err := db.MemberOfGroup(c.Request.Context(), pool, userID, groupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify membership"})
			c.Abort()
			return
		}

		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "user is not a member of the group"})
			c.Abort()
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func RequireGroupAdmin(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		groupID, ok := c.Params.Get("id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID not provided"})
			c.Abort()
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, groupID)
		if err == db.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group creator"})
			c.Abort()
			return
		}

		if creatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not the group admin"})
			c.Abort()
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func RequireGroupOwner(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		groupID, ok := c.Params.Get("id")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID not provided"})
			c.Abort()
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, groupID)
		if err == db.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			c.Abort()
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get group creator"})
			c.Abort()
			return
		}

		if creatorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not the group owner"})
			c.Abort()
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func GetGroupID(c *gin.Context) (string, bool) {
	groupIDInterface, exists := c.Get(GroupIDKey)
	if exists {
		id, ok := groupIDInterface.(string)
		if ok {
			return id, true
		}
	}

	return "", false
}

// MustGetGroupID retrieves the group ID from the context or URL parameters. Intended for use in handlers.
// If the group ID is not found, it panics, indicating a server-side misconfiguration.
func MustGetGroupID(c *gin.Context) string {
	groupID, ok := GetGroupID(c)
	if !ok {
		// not a runtime user error. Gin will recover and return 500.
		panic("MustGetGroupID: Group ID not found in context. Did you forget to add a group access middleware?")
	}
	return groupID
}
