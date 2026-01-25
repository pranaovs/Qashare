package handlers

import (
	"errors"
	"net/http"
	"slices"

	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/middleware"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupsHandler struct {
	pool *pgxpool.Pool
}

func NewGroupsHandler(pool *pgxpool.Pool) *GroupsHandler {
	return &GroupsHandler{pool: pool}
}

func (h *GroupsHandler) Create(c *gin.Context) {
	group := models.Group{}
	var err error

	group.CreatedBy = middleware.MustGetUserID(c)

	var request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group.Name, err = utils.ValidateName(request.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group.Description = request.Description
	err = db.CreateGroup(c.Request.Context(), h.pool, &group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *GroupsHandler) ListUserGroups(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	groups, err := db.MemberOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *GroupsHandler) ListAdminGroups(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groups, err := db.AdminOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (h *GroupsHandler) GetGroup(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	group, err := db.GetGroup(c.Request.Context(), h.pool, groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *GroupsHandler) AddMembers(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	type request struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := middleware.MustGetUserID(c)

	groupCreator, err := db.GetGroupCreator(c.Request.Context(), h.pool, groupID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
		return
	}
	if groupCreator != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only group admin can add members"})
		return
	}

	validUserIDs := make([]string, 0, len(req.UserIDs))
	for _, uid := range req.UserIDs {
		err := db.UserExists(c.Request.Context(), h.pool, uid)
		if err == nil {
			validUserIDs = append(validUserIDs, uid)
		} else if errors.Is(err, db.ErrUserNotFound) {
			continue
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if len(validUserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid user IDs"})
		return
	}

	err = db.AddGroupMembers(c.Request.Context(), h.pool, groupID, validUserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "members added successfully",
		"added_members": validUserIDs,
	})
}

func (h *GroupsHandler) RemoveMembers(c *gin.Context) {
	type request struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	if slices.Contains(req.UserIDs, userID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot remove group admin"})
		return
	}

	err := db.RemoveGroupMembers(c.Request.Context(), h.pool, groupID, req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "members removed",
		"removed_members": req.UserIDs,
	})
}

//
// // ListExpenses lists all expenses for a given group.
// func (h *GroupsHandler) ListExpenses(c *gin.Context) {
// 	groupID := middleware.GetGroupID(c)
//
// 	group, err := db.GetGroup(c.Request.Context(), h.pool, groupID)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		return
// 	}
//
// 	c.JSON(http.StatusOK, group)
// }
