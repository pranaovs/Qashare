package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/routes/apierrors"
	"github.com/pranaovs/qashare/utils"
)

const (
	ExpenseIDKey = "expenseID"
	ExpenseKey   = "expense"
)

// VerifyExpenseAccess checks if the authenticated user has access to the expense specified in the URL parameter "id".
// User has access if they are a member of the expense's group.
// Sets expenseID, groupID, and the expense object itself in context to avoid double-fetching.
func VerifyExpenseAccess(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		expenseID := c.Param("id")
		if expenseID == "" {
			utils.SendAbort(c, http.StatusBadRequest, "Expense ID not provided")
			return
		}

		// Get the expense to find its group
		expense, err := db.GetExpense(c.Request.Context(), pool, expenseID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrExpenseNotFound.HTTPCode, apierrors.ErrExpenseNotFound.Message)
				return
			}
			utils.SendAbort(c, http.StatusInternalServerError, "internal server error")
			return
		}

		// Check if user is a member of the expense's group
		isMember, err := db.MemberOfGroup(c.Request.Context(), pool, userID, expense.GroupID)
		if err != nil {
			utils.SendAbort(c, http.StatusInternalServerError, "failed to verify membership")
			return
		}

		if !isMember {
			utils.SendAbort(c, http.StatusForbidden, "access denied")
			return
		}

		// Cache the expense in context to avoid double-fetching
		c.Set(ExpenseKey, expense)
		c.Set(ExpenseIDKey, expenseID)
		c.Set(GroupIDKey, expense.GroupID)
		c.Next()
	}
}

// VerifyExpenseAdmin checks if the authenticated user has admin access to the expense specified in the URL parameter "id".
// A user has admin access if they are the creator of the expense's group or the creator of the expense itself.
// Sets expenseID and the expense object itself in context to avoid double-fetching.
func VerifyExpenseAdmin(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		expenseID := c.Param("id")
		if expenseID == "" {
			utils.SendAbort(c, http.StatusBadRequest, "Expense ID not provided")
			return
		}

		// Get the expense to find its group
		expense, err := db.GetExpense(c.Request.Context(), pool, expenseID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrExpenseNotFound.HTTPCode, apierrors.ErrExpenseNotFound.Message)
				return
			}
			utils.SendAbort(c, http.StatusInternalServerError, "internal server error")
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, expense.GroupID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrGroupNotFound.HTTPCode, apierrors.ErrGroupNotFound.Message)
				return
			}
			utils.SendAbort(c, http.StatusInternalServerError, "failed to get group creator")
			return
		}

		// If the user is not the group creator or the expense creator, deny access
		if creatorID != userID && (expense.AddedBy == nil || *expense.AddedBy != userID) {
			utils.SendAbort(c, http.StatusForbidden, "access denied")
			return
		}

		c.Set(ExpenseKey, expense)
		c.Set(ExpenseIDKey, expenseID)
		c.Set(GroupIDKey, expense.GroupID)
		c.Next()
	}
}

// VerifySettlementAccess checks if the authenticated user has access to the settlement specified in the URL parameter "id".
// User has access if they are a member of the settlement's group and the expense is a settlement.
// Sets expenseID, groupID, and the expense object itself in context to avoid double-fetching.
func VerifySettlementAccess(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		settlementID := c.Param("id")
		if settlementID == "" {
			utils.SendAbort(c, http.StatusBadRequest, "Settlement ID not provided")
			return
		}

		expense, err := db.GetExpense(c.Request.Context(), pool, settlementID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrExpenseNotFound.HTTPCode, "settlement not found")
				return
			}
			utils.SendAbort(c, http.StatusInternalServerError, "internal server error")
			return
		}

		if !expense.IsSettlement {
			utils.SendAbort(c, http.StatusNotFound, "settlement not found")
			return
		}

		// Check if user is a member of the settlement's group
		isMember, err := db.MemberOfGroup(c.Request.Context(), pool, userID, expense.GroupID)
		if err != nil {
			utils.SendAbort(c, http.StatusInternalServerError, "failed to verify membership")
			return
		}

		if !isMember {
			utils.SendAbort(c, http.StatusForbidden, "access denied")
			return
		}

		c.Set(ExpenseKey, expense)
		c.Set(ExpenseIDKey, settlementID)
		c.Set(GroupIDKey, expense.GroupID)
		c.Next()
	}
}

// VerifySettlementAdmin checks if the authenticated user has admin access to the settlement specified in the URL parameter "id".
// A user has admin access if they are the payer of the settlement (is_paid=true split) or the group admin.
// The expense must be a settlement (is_settlement=true).
// Derives groupID from the expense itself (no group context required).
// Sets expenseID, expense, and groupID in context to avoid double-fetching.
func VerifySettlementAdmin(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := MustGetUserID(c)

		settlementID := c.Param("id")
		if settlementID == "" {
			utils.SendAbort(c, http.StatusBadRequest, "Settlement ID not provided")
			return
		}

		expense, err := db.GetExpense(c.Request.Context(), pool, settlementID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.SendAbort(c, apierrors.ErrExpenseNotFound.HTTPCode, "settlement not found")
				return
			}
			utils.SendAbort(c, http.StatusInternalServerError, "internal server error")
			return
		}

		if !expense.IsSettlement {
			utils.SendAbort(c, http.StatusNotFound, "settlement not found")
			return
		}

		groupID := expense.GroupID

		// Check authorization: user must be the payer (is_paid=true) or group admin
		isPayerOrAdmin := false

		for _, split := range expense.Splits {
			if split.IsPaid && split.UserID == userID {
				isPayerOrAdmin = true
				break
			}
		}

		if !isPayerOrAdmin {
			creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, groupID)
			if err != nil {
				utils.SendAbort(c, http.StatusInternalServerError, "failed to get group creator")
				return
			}
			if creatorID == userID {
				isPayerOrAdmin = true
			}
		}

		if !isPayerOrAdmin {
			utils.SendAbort(c, http.StatusForbidden, "access denied")
			return
		}

		c.Set(ExpenseKey, expense)
		c.Set(ExpenseIDKey, settlementID)
		c.Set(GroupIDKey, groupID)
		c.Next()
	}
}

func GetExpenseID(c *gin.Context) (string, bool) {
	expenseIDInterface, exists := c.Get(ExpenseIDKey)
	if exists {
		id, ok := expenseIDInterface.(string)
		if ok {
			return id, true
		}
	}

	return "", false
}

// MustGetExpenseID retrieves the expense ID from the context (set by VerifyExpenseAccess). Intended for use in handlers.
// Panics if not found, indicating a server-side misconfiguration.
func MustGetExpenseID(c *gin.Context) string {
	expenseID, ok := GetExpenseID(c)
	if !ok {
		panic("MustGetExpenseID: expense ID not found in context. Did you forget to add the VerifyExpenseAccess middleware?")
	}
	return expenseID
}

// GetExpense retrieves the expense from context (cached by VerifyExpenseAccess middleware).
func GetExpense(c *gin.Context) (models.ExpenseDetails, bool) {
	expenseInterface, exists := c.Get(ExpenseKey)
	if exists {
		expense, ok := expenseInterface.(models.ExpenseDetails)
		if ok {
			return expense, true
		}
	}
	return models.ExpenseDetails{}, false
}

// MustGetExpense retrieves the expense from context. Intended for use in handlers.
// Panics if not found, indicating a server-side misconfiguration.
func MustGetExpense(c *gin.Context) models.ExpenseDetails {
	expense, ok := GetExpense(c)
	if !ok {
		panic("MustGetExpense: expense not found in context. Did you forget to add the VerifyExpenseAccess middleware?")
	}
	return expense
}
