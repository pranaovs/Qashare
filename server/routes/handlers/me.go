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

// ListGroups godoc
// @Summary List user's groups
// @Description Get all groups the logged in user is a member of
// @Tags me
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group "Returns list of groups the user is a member of"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me/groups [get]
func (h *MeHandler) ListGroups(c *gin.Context) {
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
// @Tags me
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group "Returns list of groups the user is admin of"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/me/admin [get]
func (h *MeHandler) ListAdmin(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groups, err := db.AdminOfGroups(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}
	utils.SendJSON(c, http.StatusOK, groups)
}
