package v2

import (
	"github.com/pranaovs/qashare/config"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
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

	// TODO: Implement cursor-based pagination using db.GetExpensesPaginated
	// For now, return all expenses wrapped in paginated response format
	expenses, err := db.GetExpenses(c.Request.Context(), h.pool, groupID, userID)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendData(c, models.PaginatedResponse[models.Expense]{
		Data:       expenses,
		Pagination: models.PaginationMeta{},
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

	// TODO: Implement cursor-based pagination using db.GetSettlementsPaginated
	// For now, return all settlements wrapped in paginated response format
	history, err := db.GetSettlements(c.Request.Context(), h.pool, userID, groupID)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	settlements := make([]models.Settlement, len(history))
	for i, exp := range history {
		settlements[i] = v1.ExpenseToSettlement(exp, userID)
	}

	utils.SendData(c, models.PaginatedResponse[models.Settlement]{
		Data:       settlements,
		Pagination: models.PaginationMeta{},
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

	// TODO: Implement cursor-based pagination using db.GetUserSpendingPaginated
	// For now, return all spendings wrapped in paginated response format
	expenses, err := db.GetUserSpending(c.Request.Context(), h.pool, userID, groupID)
	if err != nil {
		utils.SendError(c, err)
		return
	}

	utils.SendData(c, models.PaginatedResponse[models.UserExpense]{
		Data:       expenses,
		Pagination: models.PaginationMeta{},
	})
}

// keep compiler happy — parseCursor and uuid will be used when pagination is implemented
var _ = parseCursor
var _ uuid.UUID
