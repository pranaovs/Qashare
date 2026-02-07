package handlers

import (
	"net/http"

	"github.com/pranaovs/qashare/apperrors"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/routes/middleware"
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

// Get godoc
// @Summary Get user by ID
// @Description Get user information by user ID (users must be related through a common group)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.User "Returns user profile information"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid user ID format"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The users are not related through any common group"
// @Failure 404 {object} apierrors.AppError "USER_NOT_FOUND: The specified user does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/users/{id} [get]
func (h *UsersHandler) Get(c *gin.Context) {
	qUserID := c.Param("id")

	userID := middleware.MustGetUserID(c)

	// Do not allow access to user data if users are not related
	related, err := db.UsersRelated(c.Request.Context(), h.pool, userID, qUserID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}
	if !related {
		utils.SendError(c, apierrors.ErrUsersNotRelated)
		return
	}

	result, err := db.GetUser(c.Request.Context(), h.pool, qUserID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
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
// @Success 200 {object} models.User "Returns user profile information matching the email"
// @Failure 400 {object} apierrors.AppError "BAD_EMAIL: Invalid email format"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 404 {object} apierrors.AppError "USER_NOT_FOUND: No user found with the specified email"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/users/search/email/{email} [get]
func (h *UsersHandler) SearchByEmail(c *gin.Context) {
	_ = middleware.MustGetUserID(c)

	email, err := utils.ValidateEmail(c.Param("email"))
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidEmail: apierrors.ErrInvalidEmail,
		}))
		return
	}
	user, err := db.GetUserFromEmail(c.Request.Context(), h.pool, email)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotFound,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, user)
}

// RegisterGuest godoc
// @Summary Register a guest user
// @Description Create a new guest user by email (requires authentication). Used to add non-registered users to groups. Name will be set to [name]@domain.tld
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{email=string} true "Guest user email"
// @Success 201 {object} models.User "Guest user successfully created"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body format or missing required fields | BAD_EMAIL: Invalid email format"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 409 {object} apierrors.AppError "EMAIL_EXISTS: An account with this email already exists"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/users/guest [post]
func (h *UsersHandler) RegisterGuest(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	email, err := utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidEmail: apierrors.ErrInvalidEmail,
		}))
		return
	}

	user, err := db.CreateGuest(c.Request.Context(), h.pool, email, userID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrDuplicateKey: apierrors.ErrEmailAlreadyExists,
		}))
		return
	}

	utils.SendJSON(c, http.StatusCreated, user)
}
