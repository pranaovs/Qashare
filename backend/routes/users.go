package routes

import (
	"context"
	"net/http"

	"shared-expenses-app/db"
	"shared-expenses-app/models"
	"shared-expenses-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterUsersRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	// BUG: REMOVE THIS ROUTE IN PRODUCTION - ONLY FOR TESTING PURPOSES
	router.GET("list", func(c *gin.Context) {
		rows, err := pool.Query(context.Background(),
			`SELECT user_id, user_name, email, is_guest, password_hash, extract(epoch from created_at)::bigint
			 FROM users ORDER BY created_at DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var u models.User
			err := rows.Scan(&u.UserID, &u.Name, &u.Email, &u.Guest, &u.PasswordHash, &u.CreatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			users = append(users, u)
		}

		c.JSON(http.StatusOK, users)
	})

	// Register a new user
	router.POST("register", func(c *gin.Context) {
		var request struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password"`
		}

		// Convert request JSON body to struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate and convert inputs
		name, err := utils.ValidateName(request.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		email, err := utils.ValidateEmail(request.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		passwordHash, err := utils.HashPassword(request.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// At this point, all inputs are valid

		// Create new user
		userID, err := db.CreateUser(context.Background(), pool, name, email, passwordHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "user registered successfully",
			"user_id": userID,
		})
	})

	// Login user
	router.POST("login", func(c *gin.Context) {
		var request struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		// Convert request JSON body to struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		email, err := utils.ValidateEmail(request.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		password := request.Password

		userID, savedPassword, err := db.GetUserCredentials(context.Background(), pool, email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		}

		if ok := utils.CheckPassword(password, savedPassword); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		// At this point, login is successful

		token, err := utils.GenerateJWT(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
			return
		}

		// Return JWT token
		c.JSON(http.StatusOK, gin.H{
			"message": "login successful",
			"token":   token,
		})
	})

	// Logged in user details
	router.GET("me", func(c *gin.Context) {
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		var user models.User

		user, err = db.GetUser(context.Background(), pool, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	})

	// User details
	router.GET(":id", func(c *gin.Context) {
		qUserID := c.Param("id")

		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}

		ok, err := db.UsersRelated(context.Background(), pool, userID, qUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		}

		// At this point, users are related

		result, err := db.GetUser(context.Background(), pool, qUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, result)
	})
}
