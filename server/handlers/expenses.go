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
// @Description Create a new expense with splits for a group. The logged in user will be set as the AddedBy user.
// @Tags expenses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ExpenseDetails true "Expense details with splits"
// @Success 201 {object} models.ExpenseDetails
// @Failure 400 {object} models.ValidationErrorResponse
// @Failure 401 {object} models.UnauthorizedErrorResponse
// @Failure 403 {object} models.ForbiddenErrorResponse
// @Failure 500 {object} models.InternalErrorResponse
// @Router /expenses/ [post]
func (h *ExpensesHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var expense models.ExpenseDetails
	if err := c.ShouldBindJSON(&expense); err != nil {
		utils.LogWarn(c.Request.Context(), "Failed to bind JSON for expense creation", "user_id", userID, "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid request body", models.ErrCodeValidation, err.Error()))
		return
	}

	expense.AddedBy = &userID

	// Verify user is a member of the group
	isMember, err := db.MemberOfGroup(c.Request.Context(), h.pool, userID, expense.GroupID)
	if err != nil {
		utils.LogError(c.Request.Context(), "Failed to verify group membership", err, "user_id", userID, "group_id", expense.GroupID)
		errResp := mapDBError(err)
		utils.SendErrorWithCode(c, http.StatusInternalServerError, errResp)
		return
	}
	if !isMember {
		utils.LogWarn(c.Request.Context(), "User not a member of group", "user_id", userID, "group_id", expense.GroupID)
		utils.SendErrorWithCode(c, http.StatusForbidden,
			models.NewSimpleErrorResponse("user not a member of group", models.ErrCodeForbidden))
		return
	}

	if len(expense.Splits) == 0 {
		utils.LogWarn(c.Request.Context(), "No splits provided for expense", "user_id", userID, "group_id", expense.GroupID)
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewSimpleErrorResponse("no splits provided", models.ErrCodeValidation))
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
		utils.LogWarn(c.Request.Context(), "Split user not in group", "user_id", userID, "group_id", expense.GroupID, "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("split user not in group", models.ErrCodeValidation, err.Error()))
		return
	}

	if !expense.IsIncompleteAmount && !expense.IsIncompleteSplit {
		tolerance, err := strconv.ParseFloat(utils.GetEnv("SPLIT_TOLERANCE", "0.01"), 64)
		if err != nil {
			tolerance = 0.01
		}
		if math.Abs(paidTotal-expense.Amount) > tolerance {
			utils.LogWarn(c.Request.Context(), "Paid split total mismatch", "user_id", userID, "group_id", expense.GroupID, "paid_total", paidTotal, "amount", expense.Amount)
			utils.SendErrorWithCode(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("paid split total does not match expense amount", models.ErrCodeValidation))
			return
		}
		if math.Abs(owedTotal-expense.Amount) > tolerance {
			utils.LogWarn(c.Request.Context(), "Owed split total mismatch", "user_id", userID, "group_id", expense.GroupID, "owed_total", owedTotal, "amount", expense.Amount)
			utils.SendErrorWithCode(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("owed split total does not match expense amount", models.ErrCodeValidation))
			return
		}
	}

	err = db.CreateExpense(c.Request.Context(), h.pool, &expense)
	if err != nil {
		utils.LogError(c.Request.Context(), "Failed to create expense", err, "user_id", userID, "group_id", expense.GroupID)
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeConflict {
			status = http.StatusConflict
		}
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.LogInfo(c.Request.Context(), "Expense created successfully", "user_id", userID, "expense_id", expense.ExpenseID, "group_id", expense.GroupID)
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
// @Failure 401 {object} models.UnauthorizedErrorResponse
// @Failure 403 {object} models.ForbiddenErrorResponse
// @Failure 404 {object} models.NotFoundErrorResponse
// @Router /expenses/{id} [get]
func (h *ExpensesHandler) GetExpense(c *gin.Context) {
	// Expense is already fetched and authorized by middleware
	expense := middleware.MustGetExpense(c)
	utils.LogInfo(c.Request.Context(), "Expense retrieved", "expense_id", expense.ExpenseID)
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
// @Failure 400 {object} models.ValidationErrorResponse
// @Failure 401 {object} models.UnauthorizedErrorResponse
// @Failure 403 {object} models.ForbiddenErrorResponse
// @Failure 404 {object} models.NotFoundErrorResponse
// @Failure 500 {object} models.InternalErrorResponse
// @Router /expenses/{id} [put]
func (h *ExpensesHandler) Update(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)
	expense := middleware.MustGetExpense(c)

	var payload models.ExpenseDetails
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.LogWarn(c.Request.Context(), "Failed to bind JSON for expense update", "expense_id", expense.ExpenseID, "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("invalid request body", models.ErrCodeValidation, err.Error()))
		return
	}

	// Do not allow changing critical fields
	payload.ExpenseID = expense.ExpenseID
	payload.GroupID = expense.GroupID
	payload.AddedBy = expense.AddedBy
	payload.CreatedAt = expense.CreatedAt

	if len(payload.Splits) == 0 {
		utils.LogWarn(c.Request.Context(), "No splits provided for expense update", "expense_id", expense.ExpenseID)
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewSimpleErrorResponse("no splits provided", models.ErrCodeValidation))
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
		utils.LogWarn(c.Request.Context(), "Split user not in group for expense update", "expense_id", expense.ExpenseID, "group_id", groupID, "error", err.Error())
		utils.SendErrorWithCode(c, http.StatusBadRequest,
			models.NewErrorResponse("split user not in group", models.ErrCodeValidation, err.Error()))
		return
	}

	if !payload.IsIncompleteAmount && !payload.IsIncompleteSplit {
		tolerance, err := strconv.ParseFloat(utils.GetEnv("SPLIT_TOLERANCE", "0.01"), 64)
		if err != nil {
			tolerance = 0.01
		}
		if math.Abs(paidTotal-payload.Amount) > tolerance {
			utils.LogWarn(c.Request.Context(), "Paid split total mismatch on update", "expense_id", expense.ExpenseID, "paid_total", paidTotal, "amount", payload.Amount)
			utils.SendErrorWithCode(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("paid split total does not match expense amount", models.ErrCodeValidation))
			return
		}
		if math.Abs(owedTotal-payload.Amount) > tolerance {
			utils.LogWarn(c.Request.Context(), "Owed split total mismatch on update", "expense_id", expense.ExpenseID, "owed_total", owedTotal, "amount", payload.Amount)
			utils.SendErrorWithCode(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("owed split total does not match expense amount", models.ErrCodeValidation))
			return
		}
	}

	if err := db.UpdateExpense(c.Request.Context(), h.pool, &payload); err != nil {
		utils.LogError(c.Request.Context(), "Failed to update expense", err, "expense_id", expense.ExpenseID)
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeNotFound {
			status = http.StatusNotFound
		} else if errResp.Code == models.ErrCodeConflict {
			status = http.StatusConflict
		}
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.LogInfo(c.Request.Context(), "Expense updated successfully", "expense_id", expense.ExpenseID)
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
// @Failure 401 {object} models.UnauthorizedErrorResponse
// @Failure 403 {object} models.ForbiddenErrorResponse
// @Failure 404 {object} models.NotFoundErrorResponse
// @Failure 500 {object} models.InternalErrorResponse
// @Router /expenses/{id} [delete]
func (h *ExpensesHandler) Delete(c *gin.Context) {
	expense := middleware.MustGetExpense(c)

	if err := db.DeleteExpense(c.Request.Context(), h.pool, expense.ExpenseID); err != nil {
		utils.LogError(c.Request.Context(), "Failed to delete expense", err, "expense_id", expense.ExpenseID)
		errResp := mapDBError(err)
		status := http.StatusInternalServerError
		if errResp.Code == models.ErrCodeNotFound {
			status = http.StatusNotFound
		}
		utils.SendErrorWithCode(c, status, errResp)
		return
	}

	utils.LogInfo(c.Request.Context(), "Expense deleted successfully", "expense_id", expense.ExpenseID)
	utils.SendJSON(c, http.StatusOK, gin.H{"message": "expense deleted"})
}
