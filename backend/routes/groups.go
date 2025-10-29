package routes

import (
	"context"
	"net/http"

	"shared-expenses-app/db"
	"shared-expenses-app/models"
	"shared-expenses-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterGroupsRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	// BUG: Remove it from production
	router.GET("list", func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT group_id, group_name, description, created_by, extract(epoch from created_at)::bigint
			 FROM groups ORDER BY created_at DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var groups []models.Group
		for rows.Next() {
			var g models.Group
			err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			groups = append(groups, g)
		}

		c.JSON(http.StatusOK, groups)
	})

	router.POST("create", func(c *gin.Context) {
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

		group, err := db.CreateGroup(context.Background(), pool, name, request.Description, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, group)
	})

	router.GET("get/:group_id", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		qGroupID := c.Param("group_id")

		// Check membership in that group
		isMember, err := db.MemberOfGroup(c, pool, userID, qGroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify membership"})
			return
		}
		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		// Retrieve group details
		group, err := db.GetGroup(c, pool, qGroupID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// Return response
		c.JSON(http.StatusOK, group)
	})

	// Add members to a group
	router.POST("add_members", func(c *gin.Context) {
		type request struct {
			GroupID string   `json:"group_id" binding:"required"`
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
		group, err := db.GetGroup(c, pool, req.GroupID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		if group.CreatedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "only group admin can add members"})
			return
		}

		// Filter valid users (existing in DB)
		validUserIDs := make([]string, 0, len(req.UserIDs))
		for _, uid := range req.UserIDs {
			exists, err := db.UserExists(c, pool, uid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if exists {
				validUserIDs = append(validUserIDs, uid)
			}
		}

		if len(validUserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no valid user IDs"})
			return
		}

		// Add members
		err = db.AddGroupMembers(c, pool, req.GroupID, validUserIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add members"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "members added successfully",
			"added_members": validUserIDs,
		})
	})

	// Add members to a group
	router.POST("remove_members", func(c *gin.Context) {
		type request struct {
			GroupID string   `json:"group_id" binding:"required"`
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
		group, err := db.GetGroup(c, pool, req.GroupID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		if group.CreatedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "only group admin can remove members"})
			return
		}
		// Remove members
		err = db.RemoveGroupMembers(c, pool, req.GroupID, req.UserIDs)
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
