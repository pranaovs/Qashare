package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/middleware"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpensesHandler struct {
	pool *pgxpool.Pool
}

func NewExpensesHandler(pool *pgxpool.Pool) *ExpensesHandler {
	return &ExpensesHandler{pool: pool}
}

func (h *ExpensesHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var expense models.ExpenseDetails
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	expense.AddedBy = userID

	if err := db.MemberOfGroup(c.Request.Context(), h.pool, userID, expense.GroupID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "user not a member of group"})
		return
	}

	if len(expense.Splits) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no splits provided"})
		return
	}

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

	uniqueUserIDs := utils.GetUniqueUserIDs(splitUserIDs)

	if err := db.AllMembersOfGroup(c.Request.Context(), h.pool, uniqueUserIDs, expense.GroupID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "split user not in group"})
		return
	}

	if !expense.IsIncompleteAmount && !expense.IsIncompleteSplit {
		tolerance, err := strconv.ParseFloat(utils.GetEnv("SPLIT_TOLERANCE", "0.01"), 64)
		if err != nil {
			tolerance = 0.01
		}
		if math.Abs(paidTotal-expense.Amount) > tolerance {
			c.JSON(http.StatusBadRequest, gin.H{"error": "paid split total does not match expense amount"})
			return
		}
		if math.Abs(owedTotal-expense.Amount) > tolerance {
			c.JSON(http.StatusBadRequest, gin.H{"error": "owed split total does not match expense amount"})
			return
		}
	}

	expenseID, err := db.CreateExpense(c.Request.Context(), h.pool, expense)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"expense_id": expenseID})
}

func (h *ExpensesHandler) GetExpense(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	expenseID := c.Param("id")
	expense, err := db.GetExpense(c.Request.Context(), h.pool, expenseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if err := db.MemberOfGroup(c.Request.Context(), h.pool, userID, expense.GroupID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (h *ExpensesHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	expenseID := c.Param("id")
	if expenseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing expense id"})
		return
	}

	var payload models.ExpenseDetails
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	payload.ExpenseID = expenseID

	exp, err := db.GetExpense(c.Request.Context(), h.pool, expenseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	groupCreator, err := db.GetGroupCreator(c.Request.Context(), h.pool, exp.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}

	if userID != exp.AddedBy && userID != groupCreator {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	if len(payload.Splits) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no splits provided"})
		return
	}

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

	if err := db.AllMembersOfGroup(c.Request.Context(), h.pool, splitUserIDs, exp.GroupID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "split user not in group"})
		return
	}

	if !payload.IsIncompleteAmount && !payload.IsIncompleteSplit {
		tolerance, err := strconv.ParseFloat(utils.GetEnv("SPLIT_TOLERANCE", "0.01"), 64)
		if err != nil {
			tolerance = 0.01
		}
		if math.Abs(paidTotal-payload.Amount) > tolerance {
			c.JSON(http.StatusBadRequest, gin.H{"error": "paid split total does not match expense amount"})
			return
		}
		if math.Abs(owedTotal-payload.Amount) > tolerance {
			c.JSON(http.StatusBadRequest, gin.H{"error": "owed split total does not match expense amount"})
			return
		}
	}

	if err := db.UpdateExpense(c.Request.Context(), h.pool, payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "expense updated"})
}

func (h *ExpensesHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	expenseID := c.Param("id")
	if expenseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing expense id"})
		return
	}

	expense, err := db.GetExpense(c.Request.Context(), h.pool, expenseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	groupCreator, err := db.GetGroupCreator(c.Request.Context(), h.pool, expense.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch group"})
		return
	}

	if userID != expense.AddedBy && userID != groupCreator {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	if err := db.DeleteExpense(c.Request.Context(), h.pool, expenseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "expense deleted"})
}
