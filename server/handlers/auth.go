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
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var request struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user := models.User{}
	var err error

	user.Name, err = utils.ValidateName(request.Name)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user.Email, err = utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	passwordHash, err := utils.HashPassword(request.Password)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}
	user.PasswordHash = &passwordHash

	err = db.CreateUser(c.Request.Context(), h.pool, &user)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

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
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var request struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	email, err := utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	password := request.Password

	userID, savedPassword, err := db.GetUserCredentials(c.Request.Context(), h.pool, email)
	if err != nil {
		utils.SendError(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if ok := utils.CheckPassword(password, savedPassword); !ok {
		utils.SendError(c, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := utils.GenerateJWT(userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to create token")
		return
	}

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
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var user models.User

	user, err := db.GetUser(c.Request.Context(), h.pool, userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
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
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/guest [post]
func (h *AuthHandler) RegisterGuest(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	email, err := utils.ValidateEmail(request.Email)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := db.CreateGuest(c.Request.Context(), h.pool, email, userID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSON(c, http.StatusCreated, user)
}
