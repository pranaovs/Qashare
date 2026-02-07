package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/utils"
)

const GroupIDKey = "groupID"

// RequireGroupMember checks if the authenticated user is a member of the group
func RequireGroupMember(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)
		groupIDStr, ok := c.Params.Get("id")

		if !ok {
			utils.SendAbort(c, http.StatusBadRequest, "Group ID not provided")
			return
		}

		groupID, err := db.ParseUUID(groupIDStr)
		if err != nil {
			utils.SendAbort(c, http.StatusBadRequest, "Invalid Group ID format")
			return
		}

		ok, err = db.MemberOfGroup(c.Request.Context(), pool, userID, groupID)
		if err != nil {
			utils.SendAbort(c, http.StatusInternalServerError, "failed to verify membership")
			return
		}

		if !ok {
			utils.SendAbort(c, http.StatusForbidden, "user is not a member of the group")
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func RequireGroupAdmin(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		groupIDStr, ok := c.Params.Get("id")
		if !ok {
			utils.SendAbort(c, http.StatusBadRequest, "Group ID not provided")
			return
		}

		groupID, err := db.ParseUUID(groupIDStr)
		if err != nil {
			utils.SendAbort(c, http.StatusBadRequest, "Invalid Group ID format")
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, groupID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrGroupNotFound.HTTPCode, apierrors.ErrGroupNotFound.Message)
				return
			}
			utils.SendAbort(c, http.StatusInternalServerError, "failed to get group creator")
			return
		}

		if creatorID == nil || *creatorID != userID {
			utils.SendAbort(c, http.StatusForbidden, "not the group admin")
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func RequireGroupOwner(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		groupIDStr, ok := c.Params.Get("id")
		if !ok {
			utils.SendAbort(c, http.StatusBadRequest, "Group ID not provided")
			return
		}

		groupID, err := db.ParseUUID(groupIDStr)
		if err != nil {
			utils.SendAbort(c, http.StatusBadRequest, "Invalid Group ID format")
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, groupID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrGroupNotFound.HTTPCode, apierrors.ErrGroupNotFound.Message)
				return
			}
			utils.SendAbort(c, http.StatusInternalServerError, "failed to get group creator")
			return
		}

		if creatorID == nil || *creatorID != userID {
			utils.SendAbort(c, http.StatusForbidden, "not the group owner")
			return
		}

		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func GetGroupID(c *gin.Context) (uuid.UUID, bool) {
	groupIDInterface, exists := c.Get(GroupIDKey)
	if exists {
		id, ok := groupIDInterface.(uuid.UUID)
		if ok {
			return id, true
		}
	}

	return uuid.Nil, false
}

// MustGetGroupID retrieves the group ID from the context. Intended for use in handlers.
// If the group ID is not found, it panics, indicating a server-side misconfiguration.
func MustGetGroupID(c *gin.Context) uuid.UUID {
	groupID, ok := GetGroupID(c)
	if !ok {
		// not a runtime user error. Gin will recover and return 500.
		panic("MustGetGroupID: Group ID not found in context. Did you forget to add a group access middleware?")
	}
	return groupID
}
