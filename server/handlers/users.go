package handlers

import (
	"net/http"

	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/middleware"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersHandler struct {
	pool *pgxpool.Pool
}

func NewUsersHandler(pool *pgxpool.Pool) *UsersHandler {
	return &UsersHandler{pool: pool}
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user information by user ID (users must be related through a common group)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [get]
func (h *UsersHandler) GetUser(c *gin.Context) {
	ctx := c.Request.Context()
	qUserID := c.Param("id")
	userID := middleware.MustGetUserID(c)

	// Do not allow access to user data if users are not related
	related, err := db.UsersRelated(ctx, h.pool, userID, qUserID)
	if err != nil {
		utils.LogError(ctx, "Failed to check user relation", err, "requester_id", userID, "target_id", qUserID)
		utils.SendErrorWithCode(c, http.StatusInternalServerError,
			models.NewSimpleErrorResponse("failed to verify user relation", models.ErrCodeInternal))
		return
	}
	if !related {
		utils.LogWarn(ctx, "Access denied: users not related", "requester_id", userID, "target_id", qUserID)
		utils.SendErrorWithCode(c, http.StatusForbidden,
			models.NewSimpleErrorResponse("access denied", models.ErrCodeForbidden))
		return
	}

	result, err := db.GetUser(ctx, h.pool, qUserID)
	if err != nil {
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeUserNotFound {
			status = http.StatusNotFound
		}
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.SendJSON(c, http.StatusOK, result)
}

// SearchByEmail godoc
// @Summary Search user by email
// @Description Find a user by their email address
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param email path string true "User Email"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/search/email/{email} [get]
func (h *UsersHandler) SearchByEmail(c *gin.Context) {
	ctx := c.Request.Context()
	_ = middleware.MustGetUserID(c)

	email, err := utils.ValidateEmail(c.Param("email"))
	if err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid email format", models.ErrCodeValidation, err.Error()))
		return
	}

	user, err := db.GetUserFromEmail(ctx, h.pool, email)
	if err != nil {
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeUserNotFound || errResp.Code == models.ErrCodeEmailNotRegistered {
			status = http.StatusNotFound
		}
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.SendJSON(c, http.StatusOK, user)
}
