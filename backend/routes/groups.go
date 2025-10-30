package routes

import (
	"errors"
	"net/http"
	"slices"

	"shared-expenses-app/db"
	"shared-expenses-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterGroupsRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	// BUG: Remove it from production
	//
	// router.GET("list", func(c *gin.Context) {
	// 	rows, err := pool.Query(c.Request.Context(),
	// 		`SELECT group_id, group_name, description, created_by, extract(epoch from created_at)::bigint
	// 		 FROM groups ORDER BY created_at DESC`)
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// 	defer rows.Close()
	//
	// 	var groups []models.Group
	// 	for rows.Next() {
	// 		var g models.Group
	// 		err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt)
	// 		if err != nil {
	// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 			return
	// 		}
	// 		groups = append(groups, g)
	// 	}
	//
	// 	c.JSON(http.StatusOK, groups)
	// })

	// Create Group
	router.POST("/", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		var request struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
		}

		// Convert request JSON body to struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate and convert inputs
		name, err := utils.ValidateName(request.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// At this point, all inputs are valid
		groupID, err := db.CreateGroup(c.Request.Context(), pool, name, request.Description, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"group_id": groupID})
	})

	// List groups the user is a member of
	router.GET("/me", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		groups, err := db.MemberOfGroups(c.Request.Context(), pool, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, groups)
	})

	// List groups the user is admin of
	router.GET("/admin", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		groups, err := db.AdminOfGroups(c.Request.Context(), pool, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, groups)
	})

	// Get group by ID
	router.GET("/:id", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		groupID := c.Param("id")

		// Check membership in that group
		err = db.MemberOfGroup(c, pool, userID, groupID)
		if err != nil {
			if errors.Is(err, db.ErrNotMember) {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify membership"})
			}
			return
		}

		// Retrieve group details
		group, err := db.GetGroup(c, pool, groupID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// Return response
		c.JSON(http.StatusOK, group)
	})

	// Add members to a group
	router.POST("/:id/members", func(c *gin.Context) {
		groupID := c.Param("id")

		type request struct {
			UserIDs []string `json:"user_ids" binding:"required,min=1"`
		}

		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// Authenticate the requester
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Ensure requester is the admin (creator) of the group
		groupCreator, err := db.GetGroupCreator(c, pool, groupID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		if groupCreator != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "only group admin can add members"})
			return
		}

		// Filter valid users (existing in DB)
		validUserIDs := make([]string, 0, len(req.UserIDs))
		for _, uid := range req.UserIDs {
			err := db.UserExists(c, pool, uid)
			if err == nil {
				// User exists
				validUserIDs = append(validUserIDs, uid)
			} else if errors.Is(err, db.ErrUserNotFound) {
				// User doesn't exist, skip
				continue
			} else {
				// Database error
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if len(validUserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no valid user IDs"})
			return
		}

		// Add members
		err = db.AddGroupMembers(c, pool, groupID, validUserIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add members"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "members added successfully",
			"added_members": validUserIDs,
		})
	})

	// Remove members from a group
	router.DELETE("/:id/members", func(c *gin.Context) {
		groupID := c.Param("id")

		type request struct {
			UserIDs []string `json:"user_ids" binding:"required,min=1"`
		}

		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		// Authenticate the requester
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Ensure requester is the admin (creator) of the group
		groupCreator, err := db.GetGroupCreator(c, pool, groupID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		if groupCreator != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "only group admin can remove members"})
			return
		}
		if slices.Contains(req.UserIDs, groupCreator) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove group admin"})
			return
		}

		// Remove members
		err = db.RemoveGroupMembers(c, pool, groupID, req.UserIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove members"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":         "members removed",
			"removed_members": req.UserIDs,
		})
	})
}
