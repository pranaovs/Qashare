package middleware

import (
	"github.com/gin-gonic/gin"
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
		groupID, ok := c.Params.Get("id")

		if !ok {
			utils.SendAbort(c, apierrors.ErrBadRequest.Msg("group ID not provided"))
			return
		}

		ok, err := db.MemberOfGroup(c.Request.Context(), pool, userID, groupID)
		if err != nil {
			utils.SendAbort(c, apierrors.ErrInternalServer)
			return
		}

		if !ok {
			utils.SendAbort(c, apierrors.ErrUsersNotRelated)
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
			utils.SendAbort(c, apierrors.ErrBadRequest.Msg("group ID not provided"))
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, groupID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrGroupNotFound)
				return
			}
			utils.SendAbort(c, apierrors.ErrInternalServer)
			return
		}

		if creatorID != userID {
			utils.SendAbort(c, apierrors.ErrNoPermissions.Msg("not the group admin"))
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
			utils.SendAbort(c, apierrors.ErrBadRequest.Msg("group ID not provided"))
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, groupID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrGroupNotFound)
				return
			}
			utils.SendAbort(c, apierrors.ErrInternalServer)
			return
		}

		if creatorID != userID {
			utils.SendAbort(c, apierrors.ErrNoPermissions.Msg("not the group owner"))
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
