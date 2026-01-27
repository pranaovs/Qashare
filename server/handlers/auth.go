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

type AuthHandler struct {
	pool *pgxpool.Pool
}

func NewAuthHandler(pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{pool: pool}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{name=string,email=string,password=string} true "User registration details"
// @Success 201 {object} models.User
// @Failure 400 {object} models.ErrBadRequest
// @Failure 409 {object} models.ErrConflict
// @Failure 500 {object} models.ErrInternalServer
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	utils.LogInfo(ctx, "User registration attempt", "path", c.Request.URL.Path)

	var request struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid request body", models.ErrCodeValidation, err.Error()))
		return
	}

	user := models.User{}
	var err error

	user.Name, err = utils.ValidateName(request.Name)
	if err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid name", models.ErrCodeValidation, err.Error()))
		return
	}

	user.Email, err = utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid email format", models.ErrCodeValidation, err.Error()))
		return
	}

	passwordHash, err := utils.HashPassword(request.Password)
	if err != nil {
		utils.LogError(ctx, "Failed to hash password", err)
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid password", models.ErrCodeValidation, "password must not be empty"))
		return
	}
	user.PasswordHash = &passwordHash

	err = db.CreateUser(ctx, h.pool, &user)
	if err != nil {
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeEmailExists {
			status = http.StatusConflict
		}
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.LogInfo(ctx, "User registered successfully", "user_id", user.UserID, "email", user.Email)
	utils.SendJSON(c, http.StatusCreated, user)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string,password=string} true "User login credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrBadRequest
// @Failure 401 {object} models.ErrUnauthorized
// @Failure 500 {object} models.ErrInternalServer
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	utils.LogInfo(ctx, "User login attempt", "path", c.Request.URL.Path)

	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid request body", models.ErrCodeValidation, err.Error()))
		return
	}

	email, err := utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid email format", models.ErrCodeValidation, err.Error()))
		return
	}

	password := request.Password

	userID, savedPassword, err := db.GetUserCredentials(ctx, h.pool, email)
	if err != nil {
		utils.LogError(ctx, "Failed to get user credentials", err, "email", email)
		utils.SendErrorWithCode(c, http.StatusUnauthorized,
			models.NewSimpleErrorResponse("invalid email or password", models.ErrCodeInvalidCredentials))
		return
	}

	if ok := utils.CheckPassword(password, savedPassword); !ok {
		utils.LogWarn(ctx, "Invalid password attempt", "email", email)
		utils.SendErrorWithCode(c, http.StatusUnauthorized,
			models.NewSimpleErrorResponse("invalid email or password", models.ErrCodeInvalidCredentials))
		return
	}

	token, err := utils.GenerateJWT(userID)
	if err != nil {
		utils.LogError(ctx, "Failed to generate JWT", err, "user_id", userID)
		utils.SendErrorWithCode(c, http.StatusInternalServerError,
			models.NewSimpleErrorResponse("failed to create token", models.ErrCodeInternal))
		return
	}

	utils.LogInfo(ctx, "User logged in successfully", "user_id", userID)
	utils.SendJSON(c, http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
	})
}

// Me godoc
// @Summary Get current user
// @Description Get the authenticated user's profile information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} models.ErrUnauthorized
// @Failure 404 {object} models.ErrNotFound
// @Failure 500 {object} models.ErrInternalServer
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.MustGetUserID(c)

	var user models.User

	user, err := db.GetUser(ctx, h.pool, userID)
	if err != nil {
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeUserNotFound {
			status = http.StatusNotFound
		}
		utils.SendErrorWithCode(c, status, errResp)
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
// @Param request body object{email=string} true "Guest user email"
// @Success 201 {object} models.User
// @Failure 400 {object} models.ErrBadRequest
// @Failure 401 {object} models.ErrUnauthorized
// @Failure 409 {object} models.ErrConflict
// @Failure 500 {object} models.ErrInternalServer
// @Router /auth/guest [post]
func (h *AuthHandler) RegisterGuest(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.MustGetUserID(c)
	utils.LogInfo(ctx, "Guest user registration attempt", "requester_id", userID)

	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid request body", models.ErrCodeValidation, err.Error()))
		return
	}

	email, err := utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid email format", models.ErrCodeValidation, err.Error()))
		return
	}

	user, err := db.CreateGuest(ctx, h.pool, email, userID)
	if err != nil {
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeEmailExists {
			status = http.StatusConflict
		}
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.LogInfo(ctx, "Guest user registered successfully", "guest_id", user.UserID, "email", user.Email, "requester_id", userID)
	utils.SendJSON(c, http.StatusCreated, user)
}
