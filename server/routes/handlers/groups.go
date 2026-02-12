package handlers

import (
	"net/http"
	"slices"

	"github.com/pranaovs/qashare/apperrors"
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/routes/middleware"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupsHandler struct {
	pool      *pgxpool.Pool
	appConfig config.AppConfig
}

func NewGroupsHandler(pool *pgxpool.Pool, appConfig config.AppConfig) *GroupsHandler {
	return &GroupsHandler{pool: pool, appConfig: appConfig}
}

// Create godoc
// @Summary Create a new group
// @Description Create a new group with the logged in user as the creator
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string,description=string} true "Group details"
// @Success 201 {object} models.GroupDetails "Group successfully created"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body format or missing required fields | BAD_NAME: Name contains invalid characters or is too short/long"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/ [post]
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
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidName: apierrors.ErrInvalidName,
		}))
		return
	}

	group.Description = request.Description
	err = db.CreateGroup(c.Request.Context(), h.pool, &group)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}

	// Fetch the created group from DB to return the complete entity with members
	created, err := db.GetGroup(c.Request.Context(), h.pool, group.GroupID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrGroupNotFound,
		}))
		return
	}

	utils.SendJSON(c, http.StatusCreated, created)
}

// ListUser godoc
// @Summary List user's groups
// @Description Get all groups the logged in user is a member of
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Deprecated
// @Success 200 {array} models.Group "Returns list of groups the user is a member of"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/me [get]
func (h *GroupsHandler) ListUser(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	groups, err := db.MemberOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}
	utils.SendJSON(c, http.StatusOK, groups)
}

// ListAdmin godoc
// @Summary List groups user administers
// @Description Get all groups that the authenticated user created (is admin of)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Deprecated
// @Success 200 {array} models.Group "Returns list of groups the user is admin of"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/admin [get]
func (h *GroupsHandler) ListAdmin(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groups, err := db.OwnerOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}
	utils.SendJSON(c, http.StatusOK, groups)
}

// Get godoc
// @Summary Get group details
// @Description Get detailed information about a group
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} models.GroupDetails "Returns group details including members and expenses"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id} [get]
func (h *GroupsHandler) Get(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	group, err := db.GetGroup(c.Request.Context(), h.pool, groupID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrGroupNotFound,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, group)
}

// Update godoc
// @Summary Update a group (full replacement)
// @Description Update group name and description (requires group admin permission). Immutable fields will be ignored if included in the request body.
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body models.Group true "Updated group details"
// @Success 200 {object} models.GroupDetails "Returns updated group"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or missing required fields"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "NO_PERMISSIONS: User is not the group admin | USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id} [put]
func (h *GroupsHandler) Update(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	var payload models.Group
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	// Strip immutable fields (silently ignore if client sends them)
	if err := utils.StripImmutableFields(&payload); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	// Validate name
	validatedName, err := utils.ValidateName(payload.Name)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidName: apierrors.ErrInvalidName,
		}))
		return
	}
	payload.Name = validatedName

	// Set immutable fields from authenticated context (no DB fetch needed)
	payload.GroupID = groupID

	err = db.UpdateGroup(c.Request.Context(), h.pool, &payload)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrGroupNotFound,
		}))
		return
	}

	// Fetch the updated group to ensure immutable fields (e.g., created_by, created_at) are correct in the response
	updatedGroup, err := db.GetGroup(c.Request.Context(), h.pool, groupID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrGroupNotFound,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, updatedGroup)
}

// Patch godoc
// @Summary Partially update a group
// @Description Update specific fields of a group. Only provided fields are updated, others remain unchanged.
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body models.GroupPatch true "Partial group details (name and/or description, all optional)"
// @Success 200 {object} models.GroupDetails "Returns updated group with all fields"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or validation failed"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "NO_PERMISSIONS: User is not the group admin | USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id} [patch]
func (h *GroupsHandler) Patch(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	var patch models.GroupPatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	current, err := db.GetGroup(c.Request.Context(), h.pool, groupID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrGroupNotFound,
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	// Validate name if provided
	if patch.Name != nil {
		validatedName, err := utils.ValidateName(*patch.Name)
		if err != nil {
			utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
				utils.ErrInvalidName: apierrors.ErrInvalidName,
			}))
			return
		}
		patch.Name = &validatedName
	}

	// Apply patch to group (only non-nil fields are applied)
	if err := utils.Patch(&current.Group, &patch); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	err = db.UpdateGroup(c.Request.Context(), h.pool, &current.Group)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrGroupNotFound,
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	// Return GroupDetails with updated Group and existing Members
	updated := models.GroupDetails{
		Group:   current.Group,
		Members: current.Members,
	}

	utils.SendJSON(c, http.StatusOK, updated)
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
// @Success 200 {object} map[string]interface{} "Returns success message and list of added member IDs"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body, missing required fields, or constraint violation"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "NO_PERMISSIONS: User is not the group admin | USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist | USER_NOT_FOUND: One or more specified users do not exist or no valid user IDs provided"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id}/members [post]
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

	// Admin permission is already verified by RequireGroupAdmin middleware

	if err := db.UsersExist(c.Request.Context(), h.pool, req.UserIDs); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}

	err := db.AddGroupMembers(c.Request.Context(), h.pool, groupID, req.UserIDs)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:            apierrors.ErrGroupNotFound,
			db.ErrConstraintViolation: apierrors.ErrBadRequest,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, gin.H{
		"message":       "members added successfully",
		"added_members": req.UserIDs,
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
// @Success 200 {object} map[string]interface{} "Returns success message and list of removed member IDs"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body, missing required fields, or attempting to remove self from group"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "NO_PERMISSIONS: User is not the group admin | USERS_NOT_RELATED: The authenticated user is not a member of the group | USER_NOT_IN_GROUP: One or more specified users are not members of the group"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id}/members [delete]
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
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotInGroup,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, gin.H{
		"message":         "members removed",
		"removed_members": req.UserIDs,
	})
}

// GetExpenses godoc
// @Summary List group expenses
// @Description Get all expenses of a group
// @Tags expenses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {array} models.Expense "Returns list of all expenses in the group"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id}/expenses [get]
func (h *GroupsHandler) GetExpenses(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)
	expenses, err := db.GetExpenses(c.Request.Context(), h.pool, groupID)
	if err != nil {
		utils.SendError(c, err) // Shouln't send any error as everything is validated in the middleware
		return
	}
	utils.SendData(c, expenses)
}

// GetSpendings godoc
// @Summary Get user expenses in group
// @Description Get all expenses where the authenticated user owes money in a specific group, with the user's owed amount per expense
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {array} models.UserExpense "List of expenses with user-specific amounts"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id}/spendings [get]
func (h *GroupsHandler) GetSpendings(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	expenses, err := db.GetUserSpending(c.Request.Context(), h.pool, userID, groupID)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendData(c, expenses)
}

// Delete godoc
// @Summary Delete a group
// @Description Delete a group and all its associated data (requires group admin/owner permission)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} map[string]string "Returns success message"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "NO_PERMISSIONS: User is not the group admin/owner"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id} [delete]
func (h *GroupsHandler) Delete(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)

	if err := db.DeleteGroup(c.Request.Context(), h.pool, groupID); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrGroupNotFound,
		}))
		return
	}

	utils.SendOK(c, "group deleted")
}
