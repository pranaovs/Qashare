package handlers

import (
	"github.com/pranaovs/qashare/apperrors"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/routes/middleware"
	"github.com/pranaovs/qashare/utils"

	"github.com/gin-gonic/gin"
)

// GetSettlements godoc
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
// @Router /v1/groups/{id}/settlements [get]
func (h *GroupsHandler) GetSettlements(c *gin.Context) {
	userID := middleware.MustGetUserID(c)
	groupID := middleware.MustGetGroupID(c)

	// Get settlements
	settlements, err := db.GetSettlements(c.Request.Context(), h.pool, userID, groupID, h.appConfig.SplitTolerance)
	if err != nil {
		utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
			db.ErrInvalidInput: apierrors.ErrBadRequest,
		}))
		return
	}

	utils.SendData(c, settlements)
}
