package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/apperrors"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/routes/middleware"
	"github.com/pranaovs/qashare/utils"
)

type MeHandler struct {
	pool *pgxpool.Pool
}

func NewMeHandler(pool *pgxpool.Pool) *MeHandler {
	return &MeHandler{pool: pool}
}

// Me godoc
// @Summary Get current user
// @Description Get the authenticated user's profile information
// @Tags me
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User "Returns the authenticated user's profile information"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 404 {object} apierrors.AppError "USER_NOT_FOUND: The authenticated user no longer exists in the database"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me [get]
func (h *MeHandler) Me(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var user models.User

	user, err := db.GetUser(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, user)
}

// GetGroups godoc
// @Summary List user's groups
// @Description Get all groups the logged in user is a member of
// @Tags me
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group "Returns list of groups the user is a member of"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me/groups [get]
func (h *MeHandler) GetGroups(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	groups, err := db.MemberOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}
	utils.SendJSON(c, http.StatusOK, groups)
}

// GetAdmin godoc
// @Summary List groups user administers
// @Description Get all groups that the authenticated user created (is admin of)
// @Tags me
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group "Returns list of groups the user is admin of"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me/admin [get]
func (h *MeHandler) GetAdmin(c *gin.Context) {
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

// Update godoc
// @Summary Update current user (full replacement)
// @Description Update the authenticated user's editable details. This is a full replacement, so all required fields (name and email) must be provided. Immutable fields will be ignored if included in the request body.
// @Tags me
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.User true "Updated user details"
// @Success 200 {object} models.User "Returns updated user"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or missing required fields"
// @Failure 400 {object} apierrors.AppError "BAD_NAME: The name provided contains invalid characters"
// @Failure 400 {object} apierrors.AppError "BAD_EMAIL: The email format is incorrect"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 404 {object} apierrors.AppError "USER_NOT_FOUND: The authenticated user no longer exists"
// @Failure 409 {object} apierrors.AppError "EMAIL_EXISTS: An account with this email already exists"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me [put]
func (h *MeHandler) Update(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var payload models.User
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	// Strip immutable fields (silently ignore if client sends them)
	if err := utils.StripImmutableFields(&payload); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	validatedName, err := utils.ValidateName(payload.Name)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidName: apierrors.ErrInvalidName,
		}))
		return
	}
	payload.Name = validatedName

	validatedEmail, err := utils.ValidateEmail(payload.Email)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidEmail: apierrors.ErrInvalidEmail,
		}))
		return
	}
	payload.Email = validatedEmail

	// Set immutable fields from authenticated context (no DB fetch needed)
	payload.UserID = userID

	err = db.UpdateUser(c.Request.Context(), h.pool, &payload)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrUserNotFound,
			db.ErrInvalidInput: apierrors.ErrBadRequest,
			db.ErrDuplicateKey: apierrors.ErrEmailAlreadyExists,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, payload)
}

// Patch godoc
// @Summary Partially update current user
// @Description Update specific fields of the authenticated user. Only provided fields are updated, others remain unchanged. Immutable fields (like user_id) will be ignored if included in the request body.
// @Tags me
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UserPatch true "Partial user details (name and/or email, all optional)"
// @Success 200 {object} models.User "Returns updated user"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or validation failed"
// @Failure 400 {object} apierrors.AppError "BAD_NAME: The name provided contains invalid characters"
// @Failure 400 {object} apierrors.AppError "BAD_EMAIL: The email format is incorrect"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 404 {object} apierrors.AppError "USER_NOT_FOUND: The authenticated user no longer exists"
// @Failure 409 {object} apierrors.AppError "EMAIL_EXISTS: An account with this email already exists"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me [patch]
func (h *MeHandler) Patch(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var patch models.UserPatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
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

	// Validate email if provided
	if patch.Email != nil {
		validatedEmail, err := utils.ValidateEmail(*patch.Email)
		if err != nil {
			utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
				utils.ErrInvalidEmail: apierrors.ErrInvalidEmail,
			}))
			return
		}
		patch.Email = &validatedEmail
	}

	current, err := db.GetUser(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrUserNotFound,
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	// Apply patch to user (only non-nil fields are applied)
	if err := utils.Patch(&current, &patch); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	err = db.UpdateUser(c.Request.Context(), h.pool, &current)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrUserNotFound,
			db.ErrInvalidInput: apierrors.ErrBadRequest,
			db.ErrDuplicateKey: apierrors.ErrEmailAlreadyExists,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, current)
}

// Delete godoc
// @Summary Delete current user account
// @Description Delete the authenticated user's account and all associated data.
// @Tags me
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Returns success message"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 404 {object} apierrors.AppError "USER_NOT_FOUND: The authenticated user no longer exists in the database"
// @Failure 409 {object} apierrors.AppError "USER_OWNS_GROUPS: User owns groups and must delete the groups or transfer ownership first"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me [delete]
func (h *MeHandler) Delete(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	// Check if user owns any groups
	ownedGroups, err := db.OwnerOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}
	if len(ownedGroups) > 0 {
		utils.SendError(c, apierrors.ErrUserOwnsGroups)
		return
	}

	if err := db.DeleteUser(c.Request.Context(), h.pool, userID); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}

	utils.SendOK(c, "account deleted")
}
