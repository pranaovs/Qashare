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
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, "Expense ID not provided")
			return
		}

		// Get the expense to find its group
		expense, err := db.GetExpense(c.Request.Context(), pool, expenseID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.AbortWithStatusJSON(c, apierrors.ErrExpenseNotFound.HTTPCode, apierrors.ErrExpenseNotFound.Message)
				return
			}
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, "internal server error")
			return
		}

		// Check if user is a member of the expense's group
		isMember, err := db.MemberOfGroup(c.Request.Context(), pool, userID, expense.GroupID)
		if err != nil {
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, "failed to verify membership")
			return
		}

		if !isMember {
			utils.AbortWithStatusJSON(c, http.StatusForbidden, "access denied")
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
			utils.AbortWithStatusJSON(c, http.StatusBadRequest, "Expense ID not provided")
			return
		}

		// Get the expense to find its group
		expense, err := db.GetExpense(c.Request.Context(), pool, expenseID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.AbortWithStatusJSON(c, apierrors.ErrExpenseNotFound.HTTPCode, apierrors.ErrExpenseNotFound.Message)
				return
			}
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, "internal server error")
			return
		}

		creatorID, err := db.GetGroupCreator(c.Request.Context(), pool, expense.GroupID)
		if err != nil {
			if db.IsNotFound(err) {
				utils.AbortWithStatusJSON(c, apierrors.ErrGroupNotFound.HTTPCode, apierrors.ErrGroupNotFound.Message)
				return
			}
			utils.AbortWithStatusJSON(c, http.StatusInternalServerError, "failed to get group creator")
			return
		}

		// If the user is not the group creator or the expense creator, deny access
		if creatorID != userID && (expense.AddedBy == nil || *expense.AddedBy != userID) {
			utils.AbortWithStatusJSON(c, http.StatusForbidden, "access denied")
			return
		}

		c.Set(ExpenseKey, expense)
		c.Set(ExpenseIDKey, expenseID)
		c.Set(GroupIDKey, expense.GroupID)
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
