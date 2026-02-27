package handlers

import (
	"math"
	"net/http"

	"github.com/google/uuid"
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

type SettlementsHandler struct {
	pool      *pgxpool.Pool
	appConfig config.AppConfig
}

func NewSettlementsHandler(pool *pgxpool.Pool, appConfig config.AppConfig) *SettlementsHandler {
	return &SettlementsHandler{pool: pool, appConfig: appConfig}
}

// GetSettle godoc
// @Summary Get payment settlements for a group
// @Description Get the payment balances between the authenticated user and all other members in a group. Positive amount means other user owes you, negative means you owe them.
// @Tags settlements
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {array} models.Settlement "List of non-zero settlement balances"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user is not a member of the specified group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error - unexpected database error"
// @Router /v1/groups/{id}/settle [get]
func (h *GroupsHandler) GetSettle(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	// Get settlements
	settlements, err := db.GetSettlement(c.Request.Context(), h.pool, userID, groupID, h.appConfig.SplitTolerance)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	utils.SendData(c, settlements)
}

// GetSettlements godoc
// @Summary Get settlement history for the current user in the group
// @Description Get all settlement transactions where the authenticated user is a participant (payer or receiver)
// @Tags settlements
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {array} models.Settlement "List of settlement history entries"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user is not a member of the specified group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/groups/{id}/settlements [get]
func (h *GroupsHandler) GetSettlements(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	history, err := db.GetSettlements(c.Request.Context(), h.pool, userID, groupID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	settlements := make([]models.Settlement, len(history))
	for i, exp := range history {
		settlements[i] = expenseToSettlement(exp, userID)
	}

	utils.SendData(c, settlements)
}

// Create godoc
// @Summary Settle a payment with another user in a group
// @Description Settle a payment with another user in a group by specifying the user_id and amount. Positive amount means you are paying them, negative means they are paying you. The settlement is stored as an expense with is_settlement=true.
// @Tags settlements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body models.Settlement true "Settle payment request"
// @Success 201 {object} models.Settlement "Created settlement expense with splits"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Cannot settle with yourself or missing group_id | INVALID_AMOUNT: Settlement amount cannot be zero"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user or the other user is not a member of the specified group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/groups/{id}/settle [post]
func (h *SettlementsHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	var req models.Settlement
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	if req.Amount == 0 {
		utils.SendError(c, apierrors.ErrInvalidAmount.Msg("settlement amount cannot be zero"))
		return
	}

	if req.UserID == userID {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("cannot settle with yourself"))
		return
	}

	// Verify other user is a member of the group
	isMember, err := db.MemberOfGroup(c.Request.Context(), h.pool, req.UserID, groupID)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}
	if !isMember {
		utils.SendError(c, apierrors.ErrUsersNotRelated.Msg("the other user is not a member of the group"))
		return
	}

	absAmount := math.Abs(req.Amount)

	// Positive amount: authenticated user pays req.UserID
	// Negative amount: req.UserID pays authenticated user
	payerID := userID
	receiverID := req.UserID
	if req.Amount < 0 {
		payerID = req.UserID
		receiverID = userID
	}

	expense := models.ExpenseDetails{
		Expense: models.Expense{
			Title:        "Settlement",
			GroupID:      groupID,
			AddedBy:      userID,
			Amount:       absAmount,
			IsSettlement: true,
		},
		Splits: []models.ExpenseSplit{
			{UserID: payerID, Amount: absAmount, IsPaid: true},
			{UserID: receiverID, Amount: absAmount, IsPaid: false},
		},
	}

	if err := db.CreateExpense(c.Request.Context(), h.pool, &expense); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	utils.SendJSON(c, http.StatusCreated, expenseToSettlement(expense, userID))
}

// expenseToSettlement converts an ExpenseDetails to a Settlement response.
// Amount sign is relative to the given userID:
//   - Positive: userID was the payer (is_paid=true) — userID paid/is owed by the other user
//   - Negative: userID was the receiver (is_paid=false) — the other user paid/is owed by userID
func expenseToSettlement(expense models.ExpenseDetails, userID uuid.UUID) models.Settlement {
	if len(expense.Splits) < 2 {
		return models.Settlement{
			CreatedAt:    expense.CreatedAt,
			TransactedAt: expense.TransactedAt,
			GroupID:      expense.GroupID,
		}
	}

	var otherUserID uuid.UUID
	var absAmount float64
	var userIsPayer bool

	for _, split := range expense.Splits {
		if split.UserID == userID {
			userIsPayer = split.IsPaid
			absAmount = split.Amount
		} else {
			otherUserID = split.UserID
		}
	}

	amount := absAmount
	if !userIsPayer {
		amount = -absAmount
	}

	return models.Settlement{
		CreatedAt:    expense.CreatedAt,
		TransactedAt: expense.TransactedAt,
		GroupID:      expense.GroupID,
		UserID:       otherUserID,
		Amount:       amount,
	}
}

// Get godoc
// @Summary Get settlement details
// @Description Get detailed information about a settlement including splits
// @Tags settlements
// @Produce json
// @Security BearerAuth
// @Param id path string true "Settlement ID"
// @Success 200 {object} models.Settlement "Returns settlement details"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "Access denied: user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "Settlement not found or expense is not a settlement"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/settlements/{id} [get]
func (h *SettlementsHandler) Get(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	expense := middleware.MustGetExpense(c)
	utils.SendData(c, expenseToSettlement(expense, userID))
}

// Update godoc
// @Summary Update a settlement
// @Description Replace a settlement with new values (requires being the payer). The user_id and settlement direction (payer/receiver) are immutable and cannot be changed. Amount must preserve the original sign convention.
// @Tags settlements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Settlement ID"
// @Param request body models.Settlement true "Updated settlement details"
// @Success 200 {object} models.Settlement "Returns updated settlement"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or cannot settle with yourself | INVALID_AMOUNT: Settlement amount cannot be zero"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "Access denied: user is not the payer | USERS_NOT_RELATED: The other user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "Settlement not found or expense is not a settlement"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/settlements/{id} [put]
func (h *SettlementsHandler) Update(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)
	expense := middleware.MustGetExpense(c)

	var req models.Settlement
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	if req.Amount == 0 {
		utils.SendError(c, apierrors.ErrInvalidAmount.Msg("settlement amount cannot be zero"))
		return
	}

	// Extract current participants from existing splits
	var currentPayerID, currentReceiverID uuid.UUID
	for _, split := range expense.Splits {
		if split.IsPaid {
			currentPayerID = split.UserID
		} else {
			currentReceiverID = split.UserID
		}
	}

	// Determine the existing other user (non-authenticated participant)
	existingOtherUserID := currentReceiverID
	if currentReceiverID == userID {
		existingOtherUserID = currentPayerID
	}

	// UserID is immutable: reject if client tries to change the other party
	if req.UserID != existingOtherUserID {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("settlement user_id cannot be changed"))
		return
	}

	// Direction is immutable: reject if amount sign would flip payer/receiver
	currentDirectionPositive := currentPayerID == userID
	if (req.Amount > 0) != currentDirectionPositive {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("settlement direction cannot be changed"))
		return
	}

	absAmount := math.Abs(req.Amount)

	// Preserve existing direction (payer/receiver are immutable)
	payerID := currentPayerID
	receiverID := currentReceiverID

	// Preserve existing transacted_at when client omits it (nil = not provided)
	transactedAt := req.TransactedAt
	if transactedAt == nil {
		transactedAt = expense.TransactedAt
	}

	updated := models.ExpenseDetails{
		Expense: models.Expense{
			GroupID:      groupID,
			AddedBy:      expense.AddedBy,
			TransactedAt: transactedAt,
			Amount:       absAmount,
			IsSettlement: true,
		},
		Splits: []models.ExpenseSplit{
			{UserID: payerID, Amount: absAmount, IsPaid: true},
			{UserID: receiverID, Amount: absAmount, IsPaid: false},
		},
	}

	utils.RestoreImmutableFields(&updated.Expense, &expense.Expense)

	if err := db.UpdateExpense(c.Request.Context(), h.pool, &updated); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrExpenseNotFound,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, expenseToSettlement(updated, userID))
}

// Patch godoc
// @Summary Partially update a settlement
// @Description Update specific fields of a settlement (requires being the payer). Only provided fields are updated. The user_id and settlement direction (payer/receiver) are immutable and cannot be changed. If amount is provided, its sign must preserve the original direction.
// @Tags settlements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Settlement ID"
// @Param request body models.SettlementPatch true "Partial settlement details (all fields optional)"
// @Success 200 {object} models.Settlement "Returns updated settlement"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or cannot settle with yourself | INVALID_AMOUNT: Settlement amount cannot be zero"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "Access denied: user is not the payer | USERS_NOT_RELATED: The other user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "Settlement not found or expense is not a settlement"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/settlements/{id} [patch]
func (h *SettlementsHandler) Patch(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	expense := middleware.MustGetExpense(c)

	var patch models.SettlementPatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	// Read current payer/receiver from existing splits
	var currentPayerID, currentReceiverID uuid.UUID
	for _, split := range expense.Splits {
		if split.IsPaid {
			currentPayerID = split.UserID
		} else {
			currentReceiverID = split.UserID
		}
	}

	// Validate and normalize amount before patching
	if patch.Amount != nil {
		if *patch.Amount == 0 {
			utils.SendError(c, apierrors.ErrInvalidAmount.Msg("settlement amount cannot be zero"))
			return
		}

		// Direction is immutable: reject if amount sign would flip payer/receiver
		currentDirectionPositive := currentPayerID == userID
		if (*patch.Amount > 0) != currentDirectionPositive {
			utils.SendError(c, apierrors.ErrBadRequest.Msg("settlement direction cannot be changed"))
			return
		}

		// Normalize to absolute value (Expense.Amount is always positive)
		absAmount := math.Abs(*patch.Amount)
		patch.Amount = &absAmount
	}

	// Apply patch to expense (only non-nil fields are applied)
	if err := utils.Patch(&expense.Expense, &patch); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	// Rebuild splits with the (potentially updated) amount
	expense.Splits = []models.ExpenseSplit{
		{UserID: currentPayerID, Amount: expense.Amount, IsPaid: true},
		{UserID: currentReceiverID, Amount: expense.Amount, IsPaid: false},
	}

	if err := db.UpdateExpense(c.Request.Context(), h.pool, &expense); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrExpenseNotFound,
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, expenseToSettlement(expense, userID))
}

// Delete godoc
// @Summary Delete a settlement
// @Description Delete a settlement (requires being the payer)
// @Tags settlements
// @Produce json
// @Security BearerAuth
// @Param id path string true "Settlement ID"
// @Success 200 {object} map[string]string "Returns success message"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "Access denied: user is not the payer"
// @Failure 404 {object} apierrors.AppError "Settlement not found or expense is not a settlement"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/settlements/{id} [delete]
func (h *SettlementsHandler) Delete(c *gin.Context) {
	expense := middleware.MustGetExpense(c)

	if err := db.DeleteExpense(c.Request.Context(), h.pool, expense.ExpenseID); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound: apierrors.ErrExpenseNotFound,
		}))
		return
	}

	utils.SendOK(c, "settlement deleted")
}
