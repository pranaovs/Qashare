package routes

import (
	"math"
	"net/http"
	"strconv"

	"shared-expenses-app/db"
	"shared-expenses-app/models"
	"shared-expenses-app/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterExpensesRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
	// Create expense with splits
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

		// Check user is in group
		if err := db.MemberOfGroup(c, pool, userID, expense.GroupID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "user not a member of group"})
			return
		}

		// Validate splits
		if len(expense.Splits) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no splits provided"})
			return
		}

		// Collect user IDs and calculate paid/owed totals
		splitUserIDs := make([]string, 0, len(expense.Splits))
		var paidTotal, owedTotal float64
		for _, s := range expense.Splits {
			splitUserIDs = append(splitUserIDs, s.UserID)
			if s.IsPaid {
				paidTotal += s.Amount
			} else {
				owedTotal += s.Amount
			}
		}

		// Get unique user IDs (same user can appear multiple times with different is_paid values)
		uniqueUserIDs := utils.GetUniqueUserIDs(splitUserIDs)

		// Check all split users are in group (single DB call)
		if err := db.AllMembersOfGroup(c, pool, uniqueUserIDs, expense.GroupID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "split user not in group"})
			return
		}

		// Skip amount validation if incomplete flags are set
		if !expense.IsIncompleteAmount && !expense.IsIncompleteSplit {
			tolerance, err := strconv.ParseFloat(utils.Getenv("SPLIT_TOLERANCE", "0.01"), 64)
			if err != nil {
				tolerance = 0.01
			}
			// Validate: paid amounts should equal expense amount
			if math.Abs(paidTotal-expense.Amount) > tolerance {
				c.JSON(http.StatusBadRequest, gin.H{"error": "paid split total does not match expense amount"})
				return
			}
			// Validate: owed amounts should equal expense amount
			if math.Abs(owedTotal-expense.Amount) > tolerance {
				c.JSON(http.StatusBadRequest, gin.H{"error": "owed split total does not match expense amount"})
				return
			}
		}

		// Create expense
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

		// Must be group member
		if err := db.MemberOfGroup(c, pool, userID, expense.GroupID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		c.JSON(http.StatusOK, expense)
	})

	// Update expense (with splits)
	router.PUT("/:id", func(c *gin.Context) {
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

		// Fetch existing expense
		exp, err := db.GetExpense(c, pool, expenseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
			return
		}

		// Get group creator to verify ownership
		groupCreator, err := db.GetGroupCreator(c, pool, exp.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
			return
		}

		// Authorization: only expense adder or group creator
		if userID != exp.AddedBy && userID != groupCreator {
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
			return
		}

		// Validate splits
		if len(payload.Splits) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no splits provided"})
			return
		}

		// Collect user IDs and calculate paid/owed totals
		splitUserIDs := make([]string, 0, len(payload.Splits))
		var paidTotal, owedTotal float64
		for _, s := range payload.Splits {
			splitUserIDs = append(splitUserIDs, s.UserID)
			if s.IsPaid {
				paidTotal += s.Amount
			} else {
				owedTotal += s.Amount
			}
		}

		// Get unique user IDs (same user can appear multiple times with different is_paid values)
		uniqueUserIDs := utils.GetUniqueUserIDs(splitUserIDs)

		// Check all split users are in group (single DB call)
		if err := db.AllMembersOfGroup(c, pool, uniqueUserIDs, exp.GroupID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "split user not in group"})
			return
		}

		// Skip amount validation if incomplete flags are set
		if !payload.IsIncompleteAmount && !payload.IsIncompleteSplit {
			tolerance, err := strconv.ParseFloat(utils.Getenv("SPLIT_TOLERANCE", "0.01"), 64)
			if err != nil {
				tolerance = 0.01
			}
			// Validate: paid amounts should equal expense amount
			if math.Abs(paidTotal-payload.Amount) > tolerance {
				c.JSON(http.StatusBadRequest, gin.H{"error": "paid split total does not match expense amount"})
				return
			}
			// Validate: owed amounts should equal expense amount
			if math.Abs(owedTotal-payload.Amount) > tolerance {
				c.JSON(http.StatusBadRequest, gin.H{"error": "owed split total does not match expense amount"})
				return
			}
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
			c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
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
