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

// Create godoc
// @Summary Create a new group
// @Description Create a new group with the logged in user as the creator
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,description=string} true "Group details"
// @Success 201 {object} models.Group
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
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
		errResp := models.NewErrorResponse("invalid input", models.ErrCodeValidation, err.Error())
		utils.LogWarn(c, "Invalid request body for group creation", "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest, errResp)
		return
	}

	group.Name, err = utils.ValidateName(request.Name)
	if err != nil {
		errResp := models.NewErrorResponse("invalid group name", models.ErrCodeValidation, err.Error())
		utils.LogWarn(c, "Invalid group name", "name", request.Name, "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest, errResp)
		return
	}

	group.Description = request.Description
	err = db.CreateGroup(c.Request.Context(), h.pool, &group)
	if err != nil {
		errResp := utils.MapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeConflict {
			status = http.StatusConflict
		}
		utils.LogError(c, "Failed to create group", "error", err.Error(), "user_id", group.CreatedBy)
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.LogInfo(c, "Group created successfully", "group_id", group.ID, "created_by", group.CreatedBy)
	utils.SendJSON(c, http.StatusCreated, group)
}

// ListUserGroups godoc
// @Summary List user's groups
// @Description Get all groups the logged in user is a member of
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /groups/me [get]
func (h *GroupsHandler) ListUserGroups(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	groups, err := db.MemberOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		errResp := utils.MapDBError(err)
		utils.LogError(c, "Failed to fetch user groups", "error", err.Error(), "user_id", userID)
		utils.SendErrorWithCode(c, http.StatusInternalServerError, errResp)
		return
	}
	utils.LogInfo(c, "Listed user groups", "user_id", userID, "count", len(groups))
	utils.SendJSON(c, http.StatusOK, groups)
}

// ListAdminGroups godoc
// @Summary List groups user administers
// @Description Get all groups that the authenticated user created (is admin of)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /groups/admin [get]
func (h *GroupsHandler) ListAdminGroups(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groups, err := db.AdminOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		errResp := utils.MapDBError(err)
		utils.LogError(c, "Failed to fetch admin groups", "error", err.Error(), "user_id", userID)
		utils.SendErrorWithCode(c, http.StatusInternalServerError, errResp)
		return
	}
	utils.LogInfo(c, "Listed admin groups", "user_id", userID, "count", len(groups))
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
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /groups/{id} [get]
func (h *GroupsHandler) GetGroup(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	group, err := db.GetGroup(c.Request.Context(), h.pool, groupID)
	if err != nil {
		errResp := utils.MapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeNotFound {
			status = http.StatusNotFound
		}
		utils.LogError(c, "Failed to get group", "error", err.Error(), "group_id", groupID)
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.LogInfo(c, "Retrieved group details", "group_id", groupID)
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
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /groups/{id}/members [post]
func (h *GroupsHandler) AddMembers(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	type request struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := models.NewErrorResponse("invalid request body", models.ErrCodeValidation, err.Error())
		utils.LogWarn(c, "Invalid request body for add members", "error", err.Error(), "group_id", groupID)
		utils.SendErrorWithCode(c, http.StatusBadRequest, errResp)
		return
	}

	userID := middleware.MustGetUserID(c)

	groupCreator, err := db.GetGroupCreator(c.Request.Context(), h.pool, groupID)
	if err != nil {
		errResp := utils.MapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeNotFound {
			status = http.StatusNotFound
		}
		utils.LogError(c, "Failed to get group creator", "error", err.Error(), "group_id", groupID)
		utils.SendErrorWithCode(c, status, errResp)
		return
	}
	if groupCreator != userID {
		errResp := models.NewErrorResponse("only group admin can add members", models.ErrCodeForbidden, "user is not the group creator")
		utils.LogWarn(c, "Unauthorized attempt to add members", "user_id", userID, "group_id", groupID)
		utils.SendErrorWithCode(c, http.StatusForbidden, errResp)
		return
	}

	validUserIDs := make([]string, 0, len(req.UserIDs))
	for _, uid := range req.UserIDs {
		err := db.UserExists(c.Request.Context(), h.pool, uid)
		if err == nil {
			validUserIDs = append(validUserIDs, uid)
		} else if errors.Is(err, db.ErrUserNotFound) {
			utils.LogWarn(c, "User not found, skipping", "user_id", uid, "group_id", groupID)
			continue
		} else {
			errResp := utils.MapDBError(err)
			utils.LogError(c, "Failed to check user existence", "error", err.Error(), "user_id", uid)
			utils.SendErrorWithCode(c, http.StatusInternalServerError, errResp)
			return
		}
	}

	if len(validUserIDs) == 0 {
		errResp := models.NewErrorResponse("no valid user IDs", models.ErrCodeValidation, "none of the provided user IDs exist")
		utils.LogWarn(c, "No valid user IDs to add", "group_id", groupID, "requested_count", len(req.UserIDs))
		utils.SendErrorWithCode(c, http.StatusBadRequest, errResp)
		return
	}

	err = db.AddGroupMembers(c.Request.Context(), h.pool, groupID, validUserIDs)
	if err != nil {
		errResp := utils.MapDBError(err)
		utils.LogError(c, "Failed to add group members", "error", err.Error(), "group_id", groupID, "user_count", len(validUserIDs))
		utils.SendErrorWithCode(c, http.StatusInternalServerError, errResp)
		return
	}

	utils.LogInfo(c, "Members added to group", "group_id", groupID, "added_count", len(validUserIDs))
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
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /groups/{id}/members [delete]
func (h *GroupsHandler) RemoveMembers(c *gin.Context) {
	type request struct {
		UserIDs []string `json:"user_ids" binding:"required,min=1"`
	}

	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := models.NewErrorResponse("invalid request body", models.ErrCodeValidation, err.Error())
		utils.LogWarn(c, "Invalid request body for remove members", "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest, errResp)
		return
	}

	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	if slices.Contains(req.UserIDs, userID) {
		errResp := models.NewErrorResponse("cannot remove group admin", models.ErrCodeValidation, "group creator cannot be removed from group")
		utils.LogWarn(c, "Attempt to remove group admin", "user_id", userID, "group_id", groupID)
		utils.SendErrorWithCode(c, http.StatusBadRequest, errResp)
		return
	}

	err := db.RemoveGroupMembers(c.Request.Context(), h.pool, groupID, req.UserIDs)
	if err != nil {
		errResp := utils.MapDBError(err)
		utils.LogError(c, "Failed to remove group members", "error", err.Error(), "group_id", groupID, "user_count", len(req.UserIDs))
		utils.SendErrorWithCode(c, http.StatusInternalServerError, errResp)
		return
	}

	utils.LogInfo(c, "Members removed from group", "group_id", groupID, "removed_count", len(req.UserIDs))
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
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /groups/{id}/expenses [get]
func (h *GroupsHandler) ListGroupExpenses(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)
	expenses, err := db.GetExpenses(c.Request.Context(), h.pool, groupID)
	if err == db.ErrInvalidInput {
		errResp := models.NewErrorResponse("invalid input", models.ErrCodeValidation, err.Error())
		utils.LogWarn(c, "Invalid input for list expenses", "group_id", groupID, "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest, errResp)
		return
	}
	if err != nil {
		errResp := utils.MapDBError(err)
		utils.LogError(c, "Failed to fetch group expenses", "error", err.Error(), "group_id", groupID)
		utils.SendErrorWithCode(c, http.StatusInternalServerError, errResp)
		return
	}
	utils.LogInfo(c, "Listed group expenses", "group_id", groupID, "count", len(expenses))
	utils.SendJSON(c, http.StatusOK, expenses)
}
