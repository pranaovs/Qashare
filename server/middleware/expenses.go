package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
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
		ctx := c.Request.Context()
		userID := MustGetUserID(c)

		expenseID := c.Param("id")
		if expenseID == "" {
			utils.LogWarn(ctx, "Expense ID not provided in request", "user_id", userID, "path", c.Request.URL.Path)
			utils.AbortWithError(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("Expense ID not provided", models.ErrCodeInvalidInput))
			return
		}

		// Get the expense to find its group
		expense, err := db.GetExpense(ctx, pool, expenseID)
		if errors.Is(err, db.ErrExpenseNotFound) {
			utils.LogWarn(ctx, "Expense not found", "expense_id", expenseID, "user_id", userID)
			utils.AbortWithError(c, http.StatusNotFound,
				models.NewSimpleErrorResponse("expense not found", models.ErrCodeExpenseNotFound))
			return
		}
		if err != nil {
			utils.LogError(ctx, "Failed to get expense", err, "expense_id", expenseID, "user_id", userID)
			utils.AbortWithError(c, http.StatusInternalServerError,
				models.NewSimpleErrorResponse("internal server error", models.ErrCodeInternal))
			return
		}

		// Check if user is a member of the expense's group
		isMember, err := db.MemberOfGroup(ctx, pool, userID, expense.GroupID)
		if err != nil {
			utils.LogError(ctx, "Failed to verify group membership", err, "user_id", userID, "group_id", expense.GroupID)
			utils.AbortWithError(c, http.StatusInternalServerError,
				models.NewSimpleErrorResponse("failed to verify membership", models.ErrCodeInternal))
			return
		}

		if !isMember {
			utils.LogWarn(ctx, "User is not a member of expense's group", "user_id", userID, "expense_id", expenseID, "group_id", expense.GroupID)
			utils.AbortWithError(c, http.StatusForbidden,
				models.NewSimpleErrorResponse("access denied", models.ErrCodeForbidden))
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
		ctx := c.Request.Context()
		userID := MustGetUserID(c)

		expenseID := c.Param("id")
		if expenseID == "" {
			utils.LogWarn(ctx, "Expense ID not provided in request", "user_id", userID, "path", c.Request.URL.Path)
			utils.AbortWithError(c, http.StatusBadRequest,
				models.NewSimpleErrorResponse("Expense ID not provided", models.ErrCodeInvalidInput))
			return
		}

		// Get the expense to find its group
		expense, err := db.GetExpense(ctx, pool, expenseID)
		if errors.Is(err, db.ErrExpenseNotFound) {
			utils.LogWarn(ctx, "Expense not found", "expense_id", expenseID, "user_id", userID)
			utils.AbortWithError(c, http.StatusNotFound,
				models.NewSimpleErrorResponse("expense not found", models.ErrCodeExpenseNotFound))
			return
		}
		if err != nil {
			utils.LogError(ctx, "Failed to get expense", err, "expense_id", expenseID, "user_id", userID)
			utils.AbortWithError(c, http.StatusInternalServerError,
				models.NewSimpleErrorResponse("internal server error", models.ErrCodeInternal))
			return
		}

		creatorID, err := db.GetGroupCreator(ctx, pool, expense.GroupID)
		if errors.Is(err, db.ErrGroupNotFound) {
			utils.LogWarn(ctx, "Group not found for expense", "group_id", expense.GroupID, "expense_id", expenseID)
			utils.AbortWithError(c, http.StatusNotFound,
				models.NewSimpleErrorResponse("group not found", models.ErrCodeGroupNotFound))
			return
		}

		if err != nil {
			utils.LogError(ctx, "Failed to get group creator", err, "group_id", expense.GroupID, "expense_id", expenseID)
			utils.AbortWithError(c, http.StatusInternalServerError,
				models.NewSimpleErrorResponse("failed to get group creator", models.ErrCodeInternal))
			return
		}

		// If the user is not the group creator or the expense creator, deny access
		if creatorID != userID && (expense.AddedBy == nil || *expense.AddedBy != userID) {
			utils.LogWarn(ctx, "User is not expense admin", "user_id", userID, "expense_id", expenseID, "group_creator", creatorID)
			utils.AbortWithError(c, http.StatusForbidden,
				models.NewSimpleErrorResponse("access denied", models.ErrCodeForbidden))
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
