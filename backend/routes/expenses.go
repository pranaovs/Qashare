package routes

import (
	"net/http"
	"shared-expenses-app/db"
	"shared-expenses-app/models"
	"shared-expenses-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterExpensesRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	// Create expense
	router.POST("/", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		var expense models.Expense
		if err := c.ShouldBindJSON(&expense); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		expense.AddedBy = userID

		if err := db.MemberOfGroup(c, pool, userID, expense.GroupID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "user not a member of group"})
			return
		}

		expenseID, err := db.CreateExpense(c, pool, expense)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"expense_id": expenseID})
	})

	// Get expense by ID
	router.GET("/:id", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		expenseID := c.Param("id")
		expense, err := db.GetExpense(c, pool, expenseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		if err := db.MemberOfGroup(c, pool, userID, expense.GroupID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		c.JSON(http.StatusOK, expense)
	})

	// Update expense
	router.PUT("/:id", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		expenseID := c.Param("id")
		if expenseID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing expense id"})
			return
		}

		var payload models.Expense
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		payload.ExpenseID = expenseID
		payload.AddedBy = userID

		exp, err := db.GetExpense(c, pool, expenseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// Get group creator to verify ownership
		groupCreator, err := db.GetGroupCreator(c, pool, exp.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
			return
		}

		// Authorization: only original adder or group owner can edit
		if userID != exp.AddedBy && userID != groupCreator {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized. only adder or owner can edit"})
			return
		}

		if err := db.UpdateExpense(c, pool, payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "expense updated"})
	})

	// Delete expense
	router.DELETE("/:id", func(c *gin.Context) {
		// Authenticate user
		userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		expenseID := c.Param("id")
		if expenseID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing expense id"})
			return
		}

		expense, err := db.GetExpense(c, pool, expenseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		// Get group creator to verify ownership
		groupCreator, err := db.GetGroupCreator(c, pool, expense.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
			return
		}

		// Authorization: only adder or owner
		if userID != expense.AddedBy && userID != groupCreator {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
			return
		}

		if err := db.DeleteExpense(c, pool, expenseID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "expense deleted"})
	})
}
