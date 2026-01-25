package handlers

import (
	"net/http"

	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/middleware"
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

func (h *UsersHandler) GetUser(c *gin.Context) {
	qUserID := c.Param("id")

	userID := middleware.MustGetUserID(c)

	// Do not allow access to user data if users are not related
	related, err := db.UsersRelated(c.Request.Context(), h.pool, userID, qUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !related {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	result, err := db.GetUser(c.Request.Context(), h.pool, qUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *UsersHandler) SearchByEmail(c *gin.Context) {
	_ = middleware.MustGetUserID(c)

	email, err := utils.ValidateEmail(c.Param("email"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}
	user, err := db.GetUserFromEmail(c.Request.Context(), h.pool, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
