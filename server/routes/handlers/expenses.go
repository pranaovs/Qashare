package handlers

import (
	"math"
	"net/http"

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

type ExpensesHandler struct {
	pool      *pgxpool.Pool
	appConfig config.AppConfig
}

func NewExpensesHandler(pool *pgxpool.Pool, appConfig config.AppConfig) *ExpensesHandler {
	return &ExpensesHandler{pool: pool, appConfig: appConfig}
}

// Create godoc
// @Summary Create a new expense
// @Description Create a new expense with splits for a group. The logged in user will be set as the AddedBy user.
// @Tags expenses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ExpenseDetails true "Expense details with splits"
// @Success 201 {object} models.ExpenseDetails "Expense successfully created with splits"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body, missing required fields, or no splits provided | INVALID_SPLIT: Split totals do not match expense amount or split validation failed"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user is not a member of the specified group | USER_NOT_IN_GROUP: One or more users in the splits are not members of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/expenses/ [post]
func (h *ExpensesHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var expense models.ExpenseDetails
	if err := c.ShouldBindJSON(&expense); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	expense.AddedBy = &userID

	// Verify user is a member of the group
	isMember, err := db.MemberOfGroup(c.Request.Context(), h.pool, userID, expense.GroupID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrGroupNotFound,
		}))
		return
	}
	if !isMember {
		utils.SendError(c, apierrors.ErrUsersNotRelated)
		return
	}

	if len(expense.Splits) == 0 {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("no splits provided"))
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
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotInGroup,
		}))
		return
	}

	if !expense.IsIncompleteAmount && !expense.IsIncompleteSplit {
		if math.Abs(paidTotal-expense.Amount) > h.appConfig.SplitTolerance {
			utils.SendError(c, apierrors.ErrInvalidSplit.Msg("paid split total does not match expense amount"))
			return
		}
		if math.Abs(owedTotal-expense.Amount) > h.appConfig.SplitTolerance {
			utils.SendError(c, apierrors.ErrInvalidSplit.Msg("owed split total does not match expense amount"))
			return
		}
	}

	err = db.CreateExpense(c.Request.Context(), h.pool, &expense)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrGroupNotFound,
		}))
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
// @Success 200 {object} models.ExpenseDetails "Returns expense details including all splits"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user is not a member of the group this expense belongs to"
// @Failure 404 {object} apierrors.AppError "EXPENSE_NOT_FOUND: The specified expense does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/expenses/{id} [get]
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
// @Success 200 {object} map[string]string "Returns success message"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or missing required fields | INVALID_SPLIT: No splits provided or split totals do not match expense amount"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "NO_PERMISSIONS: User is not the group admin or expense creator | USERS_NOT_RELATED: The authenticated user is not a member of the group | USER_NOT_IN_GROUP: One or more users in the splits are not members of the group"
// @Failure 404 {object} apierrors.AppError "EXPENSE_NOT_FOUND: The specified expense does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/expenses/{id} [put]
func (h *ExpensesHandler) Update(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)
	expense := middleware.MustGetExpense(c)

	var payload models.ExpenseDetails
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	// Do not allow changing critical fields
	payload.ExpenseID = expense.ExpenseID
	payload.GroupID = expense.GroupID
	payload.AddedBy = expense.AddedBy
	payload.CreatedAt = expense.CreatedAt

	if len(payload.Splits) == 0 {
		utils.SendError(c, apierrors.ErrInvalidSplit)
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
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrUserNotInGroup,
		}))
		return
	}

	if !payload.IsIncompleteAmount && !payload.IsIncompleteSplit {
		if math.Abs(paidTotal-payload.Amount) > h.appConfig.SplitTolerance {
			utils.SendError(c, apierrors.ErrInvalidSplit.Msg("paid split total does not match expense amount"))
			return
		}
		if math.Abs(owedTotal-payload.Amount) > h.appConfig.SplitTolerance {
			utils.SendError(c, apierrors.ErrInvalidSplit.Msg("owed split total does not match expense amount"))
			return
		}
	}

	if err := db.UpdateExpense(c.Request.Context(), h.pool, &payload); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrExpenseNotFound,
		}))
		return
	}

	utils.SendOK(c, "expense updated")
}

// Delete godoc
// @Summary Delete an expense
// @Description Delete an expense (requires being group admin or expense creator)
// @Tags expenses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Expense ID"
// @Success 200 {object} map[string]string "Returns success message"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "NO_PERMISSIONS: User is not the group admin or expense creator | USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "EXPENSE_NOT_FOUND: The specified expense does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/expenses/{id} [delete]
func (h *ExpensesHandler) Delete(c *gin.Context) {
	expense := middleware.MustGetExpense(c)

	if err := db.DeleteExpense(c.Request.Context(), h.pool, expense.ExpenseID); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrExpenseNotFound,
		}))
		return
	}

	utils.SendOK(c, "expense deleted")
}
