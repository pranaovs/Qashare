package v1

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
	appConfig config.AppConfig
	jwtConfig config.JWTConfig
}

func NewAuthHandler(pool *pgxpool.Pool, appConfig config.AppConfig, jwtConfig config.JWTConfig) *AuthHandler {
	return &AuthHandler{pool: pool, appConfig: appConfig, jwtConfig: jwtConfig}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{name=string,email=string,password=string} true "User registration details"
// @Success 202 {object} models.User "User registered, email verification required"
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

	if h.appConfig.Verification {
		user.EmailVerified = false
	} else {
		user.EmailVerified = true
	}

	verificationToken, err := db.CreateUser(c.Request.Context(), h.pool, &user, h.appConfig.VerifyEmailExpiry)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrDuplicateKey: apierrors.ErrEmailAlreadyExists,
		}))
		return
	}

	// Send verification email if verification is enabled
	if h.appConfig.Verification {
		err = utils.SendVerificationEmail(user.Email, verificationToken, h.appConfig.VerifyEmailExpiry)
		if err != nil {
			utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
				utils.ErrEmailSendFailed: apierrors.ErrInternalServer,
			}))
			return
		}
		utils.SendJSON(c, http.StatusAccepted, user)
		return
	}

	utils.SendJSON(c, http.StatusCreated, user)
}

// Verify godoc
// @Summary Verify email address
// @Description Verify a user's email address using a token sent to their email
// @Tags auth
// @Produce json
// @Param token query string true "Email verification token"
// @Success 200 {object} object{message=string} "Email successfully verified"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Missing token | EMAIL_VERIFICATION_TOKEN_ERROR: Token is invalid, malformed, or not found"
// @Failure 403 {object} apierrors.AppError "EMAIL_VERIFICATION_TOKEN_EXPIRED: Token has expired"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/verify [get]
func (h *AuthHandler) Verify(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("missing token query parameter"))
		return
	}

	token, err := uuid.Parse(tokenStr)
	if err != nil {
		utils.SendError(c, apierrors.ErrEmailVerificationTokenError)
		return
	}

	err = db.VerifyEmail(c.Request.Context(), h.pool, token)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrEmailVerificationTokenError,
			db.ErrExpiredToken: apierrors.ErrEmailVerificationTokenExpired,
		}))
		return
	}

	utils.SendOK(c, "email verified")
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
// @Failure 403 {object} apierrors.AppError "EMAIL_NOT_VERIFIED: The email address has not been verified"
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

	userID, savedPassword, emailVerified, err := db.GetUserCredentials(c.Request.Context(), h.pool, email)
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

	if h.appConfig.Verification && !emailVerified {
		utils.SendError(c, apierrors.ErrEmailNotVerified)
		return
	}

	refreshToken, tokenID, expiresAt, err := utils.GenerateRefreshToken(userID, h.jwtConfig)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	accessToken, err := utils.GenerateAccessToken(userID, tokenID, h.jwtConfig)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	err = db.StoreToken(c.Request.Context(), h.pool, tokenID, userID, expiresAt)
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
// @Summary Refresh tokens
// @Description Use a valid refresh token to get new access and refresh tokens. The old refresh token is revoked (token rotation).
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} models.TokenResponse "Returns new access and refresh tokens"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Missing refresh token | INVALID_REFRESH_TOKEN: Refresh token is invalid or already used"
// @Failure 403 {object} apierrors.AppError "EXPIRED_REFRESH_TOKEN: Refresh token has expired"
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
			utils.ErrExpiredToken: apierrors.ErrExpiredRefreshToken,
			utils.ErrInvalidToken: apierrors.ErrInvalidRefreshToken,
		}))
		return
	}

	oldTokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		utils.SendError(c, apierrors.ErrInvalidRefreshToken)
		return
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		utils.SendError(c, apierrors.ErrInvalidRefreshToken)
		return
	}

	newRefreshToken, newTokenID, newExpiresAt, err := utils.GenerateRefreshToken(userID, h.jwtConfig)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	accessToken, err := utils.GenerateAccessToken(userID, newTokenID, h.jwtConfig)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	err = db.RotateToken(c.Request.Context(), h.pool, oldTokenID, newTokenID, userID, newExpiresAt)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrInvalidRefreshToken,
		}))
		return
	}

	utils.SendData(c, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
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
// @Failure 403 {object} apierrors.AppError "EXPIRED_TOKEN: Access token has expired"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := middleware.MustGetSessionID(c)

	err := db.DeleteToken(c.Request.Context(), h.pool, sessionID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrInvalidAccessToken,
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
// @Failure 403 {object} apierrors.AppError "EXPIRED_TOKEN: Access token has expired"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	err := db.DeleteTokens(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendOK(c, "logged out from all devices")
}
