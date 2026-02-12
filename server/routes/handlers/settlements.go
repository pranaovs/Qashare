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
// @Success 200 {array} models.ExpenseDetails "List of settlement expenses with splits"
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

	utils.SendData(c, history)
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
			GroupID:      groupID,
			AddedBy:      &userID,
			Title:        req.Title,
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

	utils.SendJSON(c, http.StatusCreated, expense)
}

// expenseToSettlement converts an ExpenseDetails to a Settlement response.
// Amount sign is relative to AddedBy: positive means AddedBy paid the other user.
func expenseToSettlement(expense models.ExpenseDetails) models.Settlement {
	var amount float64
	if expense.Splits[0].IsPaid {
		amount = expense.Splits[0].Amount
	} else {
		amount = -expense.Splits[0].Amount
	}

	otherUserID := expense.Splits[0].UserID
	if otherUserID == *expense.AddedBy {
		otherUserID = expense.Splits[1].UserID
	}

	return models.Settlement{
		Title:     expense.Title,
		CreatedAt: expense.CreatedAt,
		GroupID:   expense.GroupID,
		UserID:    otherUserID,
		Amount:    amount,
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
	expense := middleware.MustGetExpense(c)
	utils.SendData(c, expenseToSettlement(expense))
}

// Update godoc
// @Summary Update a settlement
// @Description Replace a settlement with new values (requires being the payer or group admin). The user_id specifies the other party and amount specifies the settlement amount. Positive amount means you are paying them, negative means they are paying you.
// @Tags settlements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Settlement ID"
// @Param request body models.Settlement true "Updated settlement details"
// @Success 200 {object} models.Settlement "Returns updated settlement"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or cannot settle with yourself | INVALID_AMOUNT: Settlement amount cannot be zero"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "Access denied: user is not the payer or group admin | USERS_NOT_RELATED: The other user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "Settlement not found or expense is not a settlement"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/settlements/{id} [put]
func (h *SettlementsHandler) Update(c *gin.Context) {
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

	if expense.AddedBy == nil {
		// The original creator of this expense no longer exists; for now, forbid updating.
		utils.SendError(c, apierrors.ErrExpenseNotFound)
		return
	}
	addedByID := *expense.AddedBy

	if req.UserID == addedByID {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("cannot settle with yourself"))
		return
	}

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

	// Sign is relative to AddedBy (the original creator):
	//   Positive: AddedBy pays req.UserID
	//   Negative: req.UserID pays AddedBy
	payerID := addedByID
	receiverID := req.UserID
	if req.Amount < 0 {
		payerID = req.UserID
		receiverID = addedByID
	}

	updated := models.ExpenseDetails{
		Expense: models.Expense{
			GroupID:      groupID,
			AddedBy:      expense.AddedBy,
			Title:        req.Title,
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

	utils.SendJSON(c, http.StatusOK, expenseToSettlement(updated))
}

// Patch godoc
// @Summary Partially update a settlement
// @Description Update specific fields of a settlement (requires being the payer or group admin). Only provided fields are updated.
// @Tags settlements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Settlement ID"
// @Param request body models.SettlementPatch true "Partial settlement details (all fields optional)"
// @Success 200 {object} models.Settlement "Returns updated settlement"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid request body or cannot settle with yourself | INVALID_AMOUNT: Settlement amount cannot be zero"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "Access denied: user is not the payer or group admin | USERS_NOT_RELATED: The other user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "Settlement not found or expense is not a settlement"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/settlements/{id} [patch]
func (h *SettlementsHandler) Patch(c *gin.Context) {
	groupID := middleware.MustGetGroupID(c)
	expense := middleware.MustGetExpense(c)

	var patch models.SettlementPatch
	if err := c.ShouldBindJSON(&patch); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	addedByID := *expense.AddedBy

	// Apply title patch
	if patch.Title != nil {
		expense.Title = *patch.Title
	}

	// Read current payer/receiver from existing splits
	var currentPayerID, currentReceiverID string
	var currentAmount float64
	for _, split := range expense.Splits {
		if split.IsPaid {
			currentPayerID = split.UserID
			currentAmount = split.Amount
		} else {
			currentReceiverID = split.UserID
		}
	}

	// Apply patch
	newAmount := currentAmount
	if patch.Amount != nil {
		newAmount = math.Abs(*patch.Amount)
	}

	newPayerID := currentPayerID
	newReceiverID := currentReceiverID
	if patch.UserID != nil {
		// Replace the non-AddedBy participant
		if currentPayerID == addedByID {
			newReceiverID = *patch.UserID
		} else {
			newPayerID = *patch.UserID
		}

		// If amount also provided, sign determines direction relative to AddedBy
		if patch.Amount != nil && *patch.Amount < 0 {
			newPayerID = *patch.UserID
			newReceiverID = addedByID
		} else if patch.Amount != nil && *patch.Amount > 0 {
			newPayerID = addedByID
			newReceiverID = *patch.UserID
		}
	} else if patch.Amount != nil && *patch.Amount < 0 {
		// Amount-only with negative sign: set direction to "other pays AddedBy"
		otherUserID := currentReceiverID
		if currentReceiverID == addedByID {
			otherUserID = currentPayerID
		}
		newPayerID = otherUserID
		newReceiverID = addedByID
	} else if patch.Amount != nil && *patch.Amount > 0 {
		// Amount-only with positive sign: set direction to "AddedBy pays other"
		otherUserID := currentReceiverID
		if currentReceiverID == addedByID {
			otherUserID = currentPayerID
		}
		newPayerID = addedByID
		newReceiverID = otherUserID
	}

	// Validate
	if newAmount == 0 {
		utils.SendError(c, apierrors.ErrInvalidAmount.Msg("settlement amount cannot be zero"))
		return
	}

	if newPayerID == newReceiverID {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("cannot settle with yourself"))
		return
	}

	// Verify the other user (non-AddedBy participant) is a group member
	otherUserID := newReceiverID
	if newReceiverID == addedByID {
		otherUserID = newPayerID
	}
	isMember, err := db.MemberOfGroup(c.Request.Context(), h.pool, otherUserID, groupID)
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

	expense.Amount = newAmount
	expense.Splits = []models.ExpenseSplit{
		{UserID: newPayerID, Amount: newAmount, IsPaid: true},
		{UserID: newReceiverID, Amount: newAmount, IsPaid: false},
	}

	if err := db.UpdateExpense(c.Request.Context(), h.pool, &expense); err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrNotFound:     apierrors.ErrExpenseNotFound,
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	utils.SendJSON(c, http.StatusOK, expenseToSettlement(expense))
}

// Delete godoc
// @Summary Delete a settlement
// @Description Delete a settlement (requires being the payer or group admin)
// @Tags settlements
// @Produce json
// @Security BearerAuth
// @Param id path string true "Settlement ID"
// @Success 200 {object} map[string]string "Returns success message"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "Access denied: user is not the payer or group admin"
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
