// Package models defines patch DTOs for PATCH operations.
// These use pointer fields to distinguish "field not provided" (nil) from
// "field explicitly set to zero value" (non-nil pointer to zero).
package models

// UserPatch represents a partial update to a User.
// Only non-nil fields will be applied to the original.
type UserPatch struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}

// Apply applies the patch to the target User.
// Only non-nil fields in the patch are applied.
func (p *UserPatch) Apply(target *User) {
	if p.Name != nil {
		target.Name = *p.Name
	}
	if p.Email != nil {
		target.Email = *p.Email
	}
}

// GroupPatch represents a partial update to a Group.
// Only non-nil fields will be applied to the original.
type GroupPatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Apply applies the patch to the target Group.
// Only non-nil fields in the patch are applied.
func (p *GroupPatch) Apply(target *Group) {
	if p.Name != nil {
		target.Name = *p.Name
	}
	if p.Description != nil {
		target.Description = *p.Description
	}
}

// ExpensePatch represents a partial update to an Expense.
// Only non-nil fields will be applied to the original.
type ExpensePatch struct {
	Title              *string  `json:"title,omitempty"`
	Description        *string  `json:"description,omitempty"`
	Amount             *float64 `json:"amount,omitempty"`
	IsIncompleteAmount *bool    `json:"is_incomplete_amount,omitempty"`
	IsIncompleteSplit  *bool    `json:"is_incomplete_split,omitempty"`
	Latitude           *float64 `json:"latitude,omitempty"`
	Longitude          *float64 `json:"longitude,omitempty"`
}

// Apply applies the patch to the target Expense.
// Only non-nil fields in the patch are applied.
func (p *ExpensePatch) Apply(target *Expense) {
	if p.Title != nil {
		target.Title = *p.Title
	}
	if p.Description != nil {
		target.Description = p.Description
	}
	if p.Amount != nil {
		target.Amount = *p.Amount
	}
	if p.IsIncompleteAmount != nil {
		target.IsIncompleteAmount = *p.IsIncompleteAmount
	}
	if p.IsIncompleteSplit != nil {
		target.IsIncompleteSplit = *p.IsIncompleteSplit
	}
	if p.Latitude != nil {
		target.Latitude = p.Latitude
	}
	if p.Longitude != nil {
		target.Longitude = p.Longitude
	}
}

// ExpenseDetailsPatch represents a partial update to an ExpenseDetails.
// Only non-nil fields will be applied to the original.
type ExpenseDetailsPatch struct {
	ExpensePatch
	Splits *[]ExpenseSplit `json:"splits,omitempty"`
}

// Apply applies the patch to the target ExpenseDetails.
// Only non-nil fields in the patch are applied.
func (p *ExpenseDetailsPatch) Apply(target *ExpenseDetails) {
	p.ExpensePatch.Apply(&target.Expense)
	if p.Splits != nil {
		target.Splits = *p.Splits
	}
}
