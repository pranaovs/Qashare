package v2

import (
	"net/http"

	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/routes/middleware"
	v1 "github.com/pranaovs/qashare/routes/v1"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupsHandler struct {
	pool      *pgxpool.Pool
	appConfig config.AppConfig
	pageSize  int
}

func NewGroupsHandler(pool *pgxpool.Pool, appConfig config.AppConfig) *GroupsHandler {
	return &GroupsHandler{pool: pool, appConfig: appConfig, pageSize: appConfig.PaginationPageSize}
}

// GetExpenses godoc
// @Summary List group expenses (paginated)
// @Description Get expenses of a group with cursor-based pagination
// @Tags expenses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param cursor query string false "Cursor (expense_id) for next page. Omit for first page."
// @Success 200 {object} models.PaginatedResponse[models.Expense] "Paginated list of expenses"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid cursor format"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Access token is invalid"
// @Failure 403 {object} apierrors.AppError "EXPIRED_TOKEN: Access token has expired | USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v2/groups/{id}/expenses [get]
func (h *GroupsHandler) GetExpenses(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	cursor, err := parseCursor(c)
	if err != nil {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("invalid cursor format"))
		return
	}

	expenses, hasNext, err := db.GetExpensesPaginated(c.Request.Context(), h.pool, groupID, userID, cursor, h.pageSize)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	var nextCursor *uuid.UUID
	if hasNext && len(expenses) > 0 {
		last := expenses[len(expenses)-1].ExpenseID
		nextCursor = &last
	}

	utils.SendJSON(c, http.StatusOK, models.PaginatedResponse[models.Expense]{
		Data:       expenses,
		Pagination: models.PaginationMeta{NextCursor: nextCursor, HasNext: hasNext},
	})
}

// GetSettlements godoc
// @Summary Get settlement history (paginated)
// @Description Get settlement history with cursor-based pagination
// @Tags settlements
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param cursor query string false "Cursor (expense_id) for next page. Omit for first page."
// @Success 200 {object} models.PaginatedResponse[models.Settlement] "Paginated list of settlements"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid cursor format"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Access token is invalid"
// @Failure 403 {object} apierrors.AppError "EXPIRED_TOKEN: Access token has expired | USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v2/groups/{id}/settlements [get]
func (h *GroupsHandler) GetSettlements(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	cursor, err := parseCursor(c)
	if err != nil {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("invalid cursor format"))
		return
	}

	history, hasNext, err := db.GetSettlementsPaginated(c.Request.Context(), h.pool, userID, groupID, cursor, h.pageSize)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	settlements := make([]models.Settlement, len(history))
	for i, exp := range history {
		settlements[i] = v1.ExpenseToSettlement(exp, userID)
	}

	var nextCursor *uuid.UUID
	if hasNext && len(history) > 0 {
		last := history[len(history)-1].ExpenseID
		nextCursor = &last
	}

	utils.SendJSON(c, http.StatusOK, models.PaginatedResponse[models.Settlement]{
		Data:       settlements,
		Pagination: models.PaginationMeta{NextCursor: nextCursor, HasNext: hasNext},
	})
}

// GetSpendings godoc
// @Summary Get user expenses in group (paginated)
// @Description Get expenses where user owes money with cursor-based pagination
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param cursor query string false "Cursor (expense_id) for next page. Omit for first page."
// @Success 200 {object} models.PaginatedResponse[models.UserExpense] "Paginated list of user expenses"
// @Failure 400 {object} apierrors.AppError "BAD_REQUEST: Invalid cursor format"
// @Failure 401 {object} apierrors.AppError "INVALID_TOKEN: Access token is invalid"
// @Failure 403 {object} apierrors.AppError "EXPIRED_TOKEN: Access token has expired | USERS_NOT_RELATED: The authenticated user is not a member of the group"
// @Failure 404 {object} apierrors.AppError "GROUP_NOT_FOUND: The specified group does not exist"
// @Failure 500 {object} apierrors.AppError "Internal server error"
// @Router /v2/groups/{id}/spendings [get]
func (h *GroupsHandler) GetSpendings(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	cursor, err := parseCursor(c)
	if err != nil {
		utils.SendError(c, apierrors.ErrBadRequest.Msg("invalid cursor format"))
		return
	}

	expenses, hasNext, err := db.GetUserSpendingPaginated(c.Request.Context(), h.pool, userID, groupID, cursor, h.pageSize)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	var nextCursor *uuid.UUID
	if hasNext && len(expenses) > 0 {
		last := expenses[len(expenses)-1].ExpenseID
		nextCursor = &last
	}

	utils.SendJSON(c, http.StatusOK, models.PaginatedResponse[models.UserExpense]{
		Data:       expenses,
		Pagination: models.PaginationMeta{NextCursor: nextCursor, HasNext: hasNext},
	})
}
