package handlers

import (
	"net/http"

	"github.com/google/uuid"
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

type AuthHandler struct {
	pool      *pgxpool.Pool
	jwtConfig config.JWTConfig
}

func NewAuthHandler(pool *pgxpool.Pool, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{pool: pool, jwtConfig: jwtConfig}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{name=string,email=string,password=string} true "User registration details"
// @Success 201 {object} models.User "User successfully registered"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body format, missing required fields, or JSON parsing error | BAD_NAME: Name contains invalid characters or is too short/long | BAD_EMAIL: Invalid email format | BAD_PASSWORD: Password does not meet requirements (e.g., too short, too weak)"
// @Failure 409 {object} apierrors.AppError "EMAIL_EXISTS: An account with this email already exists"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database or system error"
// @Router /v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var request struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	user := models.User{}
	var err error

	user.Name, err = utils.ValidateName(request.Name)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidName: apierrors.ErrInvalidName,
		}))
		return
	}

	user.Email, err = utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidEmail: apierrors.ErrInvalidEmail,
		}))
		return
	}

	passwordHash, err := utils.HashPassword(request.Password)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrInvalidPassword: apierrors.ErrInvalidPassword,
			utils.ErrHashingFailed:   apierrors.ErrBadRequest,
		}))
		return
	}
	user.PasswordHash = &passwordHash

	err = db.CreateUser(c.Request.Context(), h.pool, &user)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrDuplicateKey: apierrors.ErrEmailAlreadyExists,
		}))
		return
	}

	utils.SendJSON(c, http.StatusCreated, user)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string,password=string} true "User login credentials"
// @Success 200 {object} models.TokenResponse "Returns access and refresh tokens"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body format or missing required fields | BAD_EMAIL: Invalid email format"
// @Failure 401 {object} apierrors.AppError "BAD_CREDENTIALS: Email or password is incorrect"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
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

	password := request.Password

	userID, savedPassword, err := db.GetUserCredentials(c.Request.Context(), h.pool, email)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrBadCredentials,
		}))
		return
	}

	if ok := utils.CheckPassword(password, savedPassword); !ok {
		utils.SendError(c, apierrors.ErrBadCredentials)
		return
	}

	refreshToken, tokenID, expiresAt, err := utils.GenerateRefreshToken(userID, h.jwtConfig)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	err = db.StoreToken(c.Request.Context(), h.pool, tokenID, userID, expiresAt)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	accessToken, err := utils.GenerateAccessToken(userID, tokenID, h.jwtConfig)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendData(c, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Use a valid refresh token to get an access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} models.TokenResponse "Returns new access token"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Missing refresh token"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Refresh token is invalid or revoked"
// @Failure 403 {object} apierrors.AppError "TOKEN_EXPIRED: Refresh token has expired"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	claims, err := utils.ExtractRefreshClaims(request.RefreshToken, h.jwtConfig)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			utils.ErrExpiredToken: apierrors.ErrExpiredToken,
			utils.ErrInvalidToken: apierrors.ErrInvalidToken,
		}))
		return
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		utils.SendError(c, apierrors.ErrInvalidToken)
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		utils.SendError(c, apierrors.ErrInvalidToken)
		return
	}

	// Check the token valid in the database (not revoked)
	valid, err := db.RefreshTokenExists(c.Request.Context(), h.pool, tokenID)
	if err != nil {
		utils.SendError(c, err)
		return
	}
	if !valid {
		utils.SendError(c, apierrors.ErrInvalidToken)
		return
	}

	accessToken, err := utils.GenerateAccessToken(userID, tokenID, h.jwtConfig)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendData(c, models.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
	})
}

// Logout godoc
// @Summary Logout current session
// @Description Revoke the refresh token associated with the current access token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{message=string} "Successfully logged out"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Access token is invalid"
// @Failure 403 {object} apierrors.AppError "TOKEN_EXPIRED: Access token has expired"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := middleware.MustGetSessionID(c)

	err := db.DeleteRefreshToken(c.Request.Context(), h.pool, sessionID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrInvalidToken,
		}))
		return
	}

	utils.SendOK(c, "logged out")
}

// LogoutAll godoc
// @Summary Logout from all devices
// @Description Revoke all refresh tokens for the authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{message=string} "All tokens successfully revoked"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Access token is invalid"
// @Failure 403 {object} apierrors.AppError "TOKEN_EXPIRED: Access token has expired"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	err := db.DeleteUserRefreshTokens(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendOK(c, "logged out from all devices")
}

// Me godoc
// @Summary Get current user
// @Description Get the authenticated user's profile information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Deprecated
// @Success 200 {object} models.User "Returns the authenticated user's profile information"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Access token is invalid"
// @Failure 403 {object} apierrors.AppError "TOKEN_EXPIRED: Access token has expired"
// @Failure 404 {object} apierrors.AppError "USER_NOT_FOUND: The authenticated user no longer exists in the database"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
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

// RegisterGuest godoc
// @Summary Register a guest user
// @Description Create a new guest user by email (requires authentication). Used to add non-registered users to groups. Name will be set to [name]@domain.tld
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Deprecated
// @Param request body object{email=string} true "Guest user email"
// @Success 201 {object} models.User "Guest user successfully created"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body format or missing required fields | BAD_EMAIL: Invalid email format"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Access token is invalid"
// @Failure 403 {object} apierrors.AppError "TOKEN_EXPIRED: Access token has expired"
// @Failure 409 {object} apierrors.AppError "EMAIL_EXISTS: An account with this email already exists"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/guest [post]
func (h *AuthHandler) RegisterGuest(c *gin.Context) {
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
