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

// Create godoc
// @Summary Create a new expense
// @Description Create a new expense with splits for a group
// @Tags expenses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ExpenseDetails true "Expense details with splits"
// @Success 201 {object} models.ExpenseDetails
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /expenses/ [post]
func (h *ExpensesHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var expense models.ExpenseDetails
	if err := c.ShouldBindJSON(&expense); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	expense.AddedBy = &userID

	// Verify user is a member of the group
	isMember, err := db.MemberOfGroup(c.Request.Context(), h.pool, userID, expense.GroupID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "failed to verify membership")
		return
	}
	if !isMember {
		utils.SendError(c, http.StatusForbidden, "user not a member of group")
		return
	}

	if len(expense.Splits) == 0 {
		utils.SendError(c, http.StatusBadRequest, "no splits provided")
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
		utils.SendError(c, http.StatusBadRequest, "split user not in group")
		return
	}

	if !expense.IsIncompleteAmount && !expense.IsIncompleteSplit {
		tolerance, err := strconv.ParseFloat(utils.GetEnv("SPLIT_TOLERANCE", "0.01"), 64)
		if err != nil {
			tolerance = 0.01
		}
		if math.Abs(paidTotal-expense.Amount) > tolerance {
			utils.SendError(c, http.StatusBadRequest, "paid split total does not match expense amount")
			return
		}
		if math.Abs(owedTotal-expense.Amount) > tolerance {
			utils.SendError(c, http.StatusBadRequest, "owed split total does not match expense amount")
			return
		}
	}

	err = db.CreateExpense(c.Request.Context(), h.pool, &expense)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSON(c, http.StatusCreated, expense)
}

// GetExpense godoc
// @Summary Get expense details
// @Description Get detailed information about an expense including splits
// @Tags expenses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Expense ID"
// @Success 200 {object} models.ExpenseDetails
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /expenses/{id} [get]
func (h *ExpensesHandler) GetExpense(c *gin.Context) {
	// Expense is already fetched and authorized by middleware
	expense := middleware.MustGetExpense(c)
	utils.SendJSON(c, http.StatusOK, expense)
}

// Update godoc
// @Summary Update an expense
// @Description Update an existing expense (requires being group admin or expense creator)
// @Tags expenses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Expense ID"
// @Param request body models.ExpenseDetails true "Updated expense details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /expenses/{id} [put]
func (h *ExpensesHandler) Update(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)
	expense := middleware.MustGetExpense(c)

	var payload models.ExpenseDetails
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Do not allow changing critical fields
	payload.ExpenseID = expense.ExpenseID
	payload.GroupID = expense.GroupID
	payload.AddedBy = expense.AddedBy
	payload.CreatedAt = expense.CreatedAt

	if len(payload.Splits) == 0 {
		utils.SendError(c, http.StatusBadRequest, "no splits provided")
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

	if err := db.AllMembersOfGroup(c.Request.Context(), h.pool, splitUserIDs, groupID); err != nil {
		utils.SendError(c, http.StatusBadRequest, "split user not in group")
		return
	}

	if !payload.IsIncompleteAmount && !payload.IsIncompleteSplit {
		tolerance, err := strconv.ParseFloat(utils.GetEnv("SPLIT_TOLERANCE", "0.01"), 64)
		if err != nil {
			tolerance = 0.01
		}
		if math.Abs(paidTotal-payload.Amount) > tolerance {
			utils.SendError(c, http.StatusBadRequest, "paid split total does not match expense amount")
			return
		}
		if math.Abs(owedTotal-payload.Amount) > tolerance {
			utils.SendError(c, http.StatusBadRequest, "owed split total does not match expense amount")
			return
		}
	}

	if err := db.UpdateExpense(c.Request.Context(), h.pool, &payload); err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSON(c, http.StatusOK, gin.H{"message": "expense updated"})
}

// Delete godoc
// @Summary Delete an expense
// @Description Delete an expense (requires being group admin or expense creator)
// @Tags expenses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Expense ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /expenses/{id} [delete]
func (h *ExpensesHandler) Delete(c *gin.Context) {
	expense := middleware.MustGetExpense(c)

	if err := db.DeleteExpense(c.Request.Context(), h.pool, expense.ExpenseID); err != nil {
		utils.SendError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendJSON(c, http.StatusOK, gin.H{"message": "expense deleted"})
}
