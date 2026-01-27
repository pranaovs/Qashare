package handlers

import (
	"net/http"
	"slices"

	"github.com/pranaovs/qashare/apierrors"
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

// Create godoc
// @Summary Create a new group
// @Description Create a new group with the logged in user as the creator
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,description=string} true "Group details"
// @Success 201 {object} models.Group
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/ [post]
func (h *GroupsHandler) Create(c *gin.Context) {
	group := models.Group{}
	var err error

	group.CreatedBy = middleware.MustGetUserID(c)

	var request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	group.Name, err = utils.ValidateName(request.Name)
	if err != nil {
		utils.SendError(c, apierrors.ErrInvalidName)
		return
	}

	group.Description = request.Description
	err = db.CreateGroup(c.Request.Context(), h.pool, &group)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendJSON(c, http.StatusCreated, group)
}

// ListUserGroups godoc
// @Summary List user's groups
// @Description Get all groups the logged in user is a member of
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/me [get]
func (h *GroupsHandler) ListUserGroups(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	groups, err := db.MemberOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}
	utils.SendJSON(c, http.StatusOK, groups)
}

// ListAdminGroups godoc
// @Summary List groups user administers
// @Description Get all groups that the authenticated user created (is admin of)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/admin [get]
func (h *GroupsHandler) ListAdminGroups(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groups, err := db.AdminOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}
	utils.SendJSON(c, http.StatusOK, groups)
}

// GetGroup godoc
// @Summary Get group details
// @Description Get detailed information about a group
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} models.GroupDetails
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /groups/{id} [get]
func (h *GroupsHandler) GetGroup(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	group, err := db.GetGroup(c.Request.Context(), h.pool, groupID)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendJSON(c, http.StatusOK, group)
}

// AddMembers godoc
// @Summary Add members to group
// @Description Add one or more users to a group (requires group admin permission)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body object{user_ids=[]string} true "User IDs to add"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id}/members [post]
func (h *GroupsHandler) AddMembers(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	type request struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	userID := middleware.MustGetUserID(c)

	groupCreator, err := db.GetGroupCreator(c.Request.Context(), h.pool, groupID)
	if err != nil {
		utils.SendError(c, apierrors.ErrGroupNotFound)
		return
	}
	if groupCreator != userID {
		utils.SendError(c, apierrors.ErrNoPermissions)
		return
	}

	validUserIDs := make([]string, 0, len(req.UserIDs))
	for _, uid := range req.UserIDs {
		err := db.UserExists(c.Request.Context(), h.pool, uid)
		if err == nil {
			validUserIDs = append(validUserIDs, uid)
		} else {
			utils.SendError(c, err)
			return
		}
	}

	if len(validUserIDs) == 0 {
		utils.SendError(c, apierrors.ErrUserNotFound.Msg("No valid user IDs provided"))
		return
	}

	err = db.AddGroupMembers(c.Request.Context(), h.pool, groupID, validUserIDs)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendJSON(c, http.StatusOK, gin.H{
		"message":       "members added successfully",
		"added_members": validUserIDs,
	})
}

// RemoveMembers godoc
// @Summary Remove members from group
// @Description Remove one or more users from a group (requires group admin permission)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body object{user_ids=[]string} true "User IDs to remove"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id}/members [delete]
func (h *GroupsHandler) RemoveMembers(c *gin.Context) {
	type request struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	if slices.Contains(req.UserIDs, userID) {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("cannot remove self from group"))
		return
	}

	err := db.RemoveGroupMembers(c.Request.Context(), h.pool, groupID, req.UserIDs)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendJSON(c, http.StatusOK, gin.H{
		"message":         "members removed",
		"removed_members": req.UserIDs,
	})
}

// ListGroupExpenses godoc
// @Summary List group expenses
// @Description Get all expenses of a group
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {array} models.Expense
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id}/expenses [get]
func (h *GroupsHandler) ListGroupExpenses(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)
	expenses, err := db.GetExpenses(c.Request.Context(), h.pool, groupID)
	if err != nil {
		utils.SendError(c, err)
		return
	}
	utils.SendData(c, expenses)
}
