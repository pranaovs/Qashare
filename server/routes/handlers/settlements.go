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
// @Description Settle a payment with another user in a group by specifying the group_id, user_id, and amount. Positive amount means you are paying them, negative means they are paying you. The settlement is stored as an expense with is_settlement=true.
// @Tags settlements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.Settlement true "Settle payment request"
// @Success 200 {object} models.ExpenseDetails "Created settlement expense with splits"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Cannot settle with yourself or missing group_id | INVALID_AMOUNT: Settlement amount cannot be zero"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user or the other user is not a member of the specified group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/settlements/ [post]
func (h *SettlementsHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req models.Settlement
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, apierrors.ErrBadRequest)
		return
	}

	if req.GroupID == "" {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("group_id is required"))
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
	isMember, err := db.MemberOfGroup(c.Request.Context(), h.pool, req.UserID, req.GroupID)
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
			GroupID:      req.GroupID,
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

	utils.SendData(c, expense)
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

	// Settlement amount
	// Positive means the curent user paid the other, negative means the other paid the current user.
	var amount float64
	if expense.Splits[0].IsPaid {
		amount = expense.Splits[0].Amount
	} else {
		amount = -expense.Splits[0].Amount
	}

	// Find the other user (non-AddedBy participant) to include in the response
	otherUserID := expense.Splits[0].UserID
	if otherUserID == *expense.AddedBy {
		otherUserID = expense.Splits[1].UserID
	}

	payload := models.Settlement{
		Title:     expense.Title,
		CreatedAt: expense.CreatedAt,
		GroupID:   expense.GroupID,
		UserID:    otherUserID,
		Amount:    amount,
	}
	utils.SendJSON(c, http.StatusOK, payload)
}
