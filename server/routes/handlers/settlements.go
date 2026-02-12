package handlers

import (
	"math"

	"github.com/pranaovs/qashare/apperrors"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/routes/middleware"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
)

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

// Settle godoc
// @Summary Settle a payment with another user in a group
// @Description Settle a payment with another user in a group by specifying the other user's ID and the amount to settle. Positive amount means you are paying them, negative means they are paying you. The settlement is stored as an expense with is_settlement=true.
// @Tags settlements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body models.Settlement true "Settle payment request"
// @Success 200 {object} models.ExpenseDetails "Created settlement expense with splits"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Cannot settle with yourself | INVALID_AMOUNT: Settlement amount cannot be zero"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Authentication token is missing, invalid, or expired"
// @Failure 403 {object} apierrors.AppError "USERS_NOT_RELATED: The authenticated user or the other user is not a member of the specified group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v1/groups/{id}/settle [post]
func (h *GroupsHandler) Settle(c *gin.Context) {
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
			Title:        "Settlement",
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
