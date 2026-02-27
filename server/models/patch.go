// Package models defines patch DTOs for PATCH operations.
// These use pointer fields to distinguish "field not provided" (nil) from
// "field explicitly set to zero value" (non-nil pointer to zero).
//
// Use utils.Patch(target, patch) to apply these patches.
package models

// UserPatch represents a partial update to a User.
// Only non-nil fields will be applied to the target.
type UserPatch struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}

// GroupPatch represents a partial update to a Group.
// Only non-nil fields will be applied to the target.
type GroupPatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ExpensePatch represents a partial update to an Expense.
// Only non-nil fields will be applied to the target.
type ExpensePatch struct {
	Title              *string  `json:"title,omitempty"`
	Description        *string  `json:"description,omitempty"`
	TransactedAt       *int64   `json:"transacted_at,omitempty"`
	Amount             *float64 `json:"amount,omitempty"`
	IsIncompleteAmount *bool    `json:"is_incomplete_amount,omitempty"`
	IsIncompleteSplit  *bool    `json:"is_incomplete_split,omitempty"`
	Latitude           *float64 `json:"latitude,omitempty"`
	Longitude          *float64 `json:"longitude,omitempty"`
}

// ExpenseDetailsPatch represents a partial update to an ExpenseDetails.
// Only non-nil fields will be applied to the target.
type ExpenseDetailsPatch struct {
	ExpensePatch
	Splits *[]ExpenseSplit `json:"splits,omitempty"`
}

// SettlementPatch represents a partial update to a Settlement.
// Only non-nil fields will be applied to the target.
type SettlementPatch struct {
	TransactedAt *int64   `json:"transacted_at,omitempty"`
	Amount       *float64 `json:"amount,omitempty"`
}
