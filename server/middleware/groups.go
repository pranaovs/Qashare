package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"
)

const GroupIDKey = "groupID"

// RequireGroupMember checks if the authenticated user is a member of the group
func RequireGroupMember(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userID := MustGetUserID(c)
		groupID, ok := c.Params.Get("id")

		if !ok {
			utils.LogWarn(ctx, "Group ID not provided in request", "user_id", userID, "path", c.Request.URL.Path)
			utils.AbortWithError(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("Group ID not provided", models.ErrCodeInvalidInput))
			return
		}

		ok, err := db.MemberOfGroup(ctx, pool, userID, groupID)
		if err != nil {
			utils.LogError(ctx, "Failed to verify group membership", err, "user_id", userID, "group_id", groupID)
			utils.AbortWithError(c, http.StatusInternalServerError,
				models.NewSimpleErrorResponse("failed to verify membership", models.ErrCodeInternal))
			return
		}

		if !ok {
			utils.LogWarn(ctx, "User is not a member of the group", "user_id", userID, "group_id", groupID)
			utils.AbortWithError(c, http.StatusForbidden,
				models.NewSimpleErrorResponse("user is not a member of the group", models.ErrCodeNotMember))
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func RequireGroupAdmin(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userID := MustGetUserID(c)

		groupID, ok := c.Params.Get("id")
		if !ok {
			utils.LogWarn(ctx, "Group ID not provided in request", "user_id", userID, "path", c.Request.URL.Path)
			utils.AbortWithError(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("Group ID not provided", models.ErrCodeInvalidInput))
			return
		}

		creatorID, err := db.GetGroupCreator(ctx, pool, groupID)
		if errors.Is(err, db.ErrGroupNotFound) {
			utils.LogWarn(ctx, "Group not found", "group_id", groupID, "user_id", userID)
			utils.AbortWithError(c, http.StatusNotFound,
				models.NewSimpleErrorResponse("group not found", models.ErrCodeGroupNotFound))
			return
		}
		if err != nil {
			utils.LogError(ctx, "Failed to get group creator", err, "group_id", groupID, "user_id", userID)
			utils.AbortWithError(c, http.StatusInternalServerError,
				models.NewSimpleErrorResponse("failed to get group creator", models.ErrCodeInternal))
			return
		}

		if creatorID != userID {
			utils.LogWarn(ctx, "User is not the group admin", "user_id", userID, "group_id", groupID, "creator_id", creatorID)
			utils.AbortWithError(c, http.StatusForbidden,
				models.NewSimpleErrorResponse("not the group admin", models.ErrCodeNotGroupCreator))
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func RequireGroupOwner(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userID := MustGetUserID(c)

		groupID, ok := c.Params.Get("id")
		if !ok {
			utils.LogWarn(ctx, "Group ID not provided in request", "user_id", userID, "path", c.Request.URL.Path)
			utils.AbortWithError(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("Group ID not provided", models.ErrCodeInvalidInput))
			return
		}

		creatorID, err := db.GetGroupCreator(ctx, pool, groupID)
		if errors.Is(err, db.ErrGroupNotFound) {
			utils.LogWarn(ctx, "Group not found", "group_id", groupID, "user_id", userID)
			utils.AbortWithError(c, http.StatusNotFound,
				models.NewSimpleErrorResponse("group not found", models.ErrCodeGroupNotFound))
			return
		}

		if err != nil {
			utils.LogError(ctx, "Failed to get group creator", err, "group_id", groupID, "user_id", userID)
			utils.AbortWithError(c, http.StatusInternalServerError,
				models.NewSimpleErrorResponse("failed to get group creator", models.ErrCodeInternal))
			return
		}

		if creatorID != userID {
			utils.LogWarn(ctx, "User is not the group owner", "user_id", userID, "group_id", groupID, "creator_id", creatorID)
			utils.AbortWithError(c, http.StatusForbidden,
				models.NewSimpleErrorResponse("not the group owner", models.ErrCodeNotGroupCreator))
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

// MustGetGroupID retrieves the group ID from the context. Intended for use in handlers.
// If the group ID is not found, it panics, indicating a server-side misconfiguration.
func MustGetGroupID(c *gin.Context) string {
	groupID, ok := GetGroupID(c)
	if !ok {
		// not a runtime user error. Gin will recover and return 500.
		panic("MustGetGroupID: Group ID not found in context. Did you forget to add a group access middleware?")
	}
	return groupID
}
