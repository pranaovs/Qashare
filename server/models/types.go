// Package models defines the core data structures for the shared expenses application.
package models

// User represents a user in the system
type User struct {
	UserID       string  `json:"user_id" db:"user_id" immutable:"true"`
	Name         string  `json:"name" db:"user_name"`
	Email        string  `json:"email" db:"email"`
	Guest        bool    `json:"guest" db:"is_guest" immutable:"true"`
	PasswordHash *string `json:"-" db:"password_hash" immutable:"true"` // excluded from JSON responses
	CreatedAt    int64   `json:"created_at" db:"created_at" immutable:"true"`
}

// Group represents a group
type Group struct {
	GroupID     string `json:"group_id" db:"group_id" immutable:"true"`
	Name        string `json:"name" db:"group_name"`
	Description string `json:"description,omitempty" db:"description"`
	CreatedBy   string `json:"created_by" db:"created_by" immutable:"true"`
	CreatedAt   int64  `json:"created_at" db:"created_at" immutable:"true"`
}

// GroupDetails represents detailed information about a group including its members
type GroupDetails struct {
	Group               // Struct embedding to include all Group fields
	Members []GroupUser `json:"members"`
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	UserID   string `json:"user_id" db:"user_id"`
	GroupID  string `json:"group_id" db:"group_id"`
	JoinedAt int64  `json:"joined_at" db:"joined_at"`
}

// GroupUser Not a part of DB schema, used for responses
type GroupUser struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Guest    bool   `json:"guest"`
	JoinedAt int64  `json:"joined_at"`
}

// Expense represents an expense in a group(ID)
type Expense struct {
	ExpenseID          string   `json:"expense_id" db:"expense_id" immutable:"true"`
	GroupID            string   `json:"group_id" db:"group_id" immutable:"true"`
	AddedBy            *string  `json:"added_by" db:"added_by" immutable:"true"`
	Title              string   `json:"title" db:"title"`
	Description        *string  `json:"description,omitempty" db:"description"` // pointer because nullable in db
	CreatedAt          int64    `json:"created_at" db:"created_at" immutable:"true"`
	Amount             float64  `json:"amount" db:"amount"`
	IsIncompleteAmount bool     `json:"is_incomplete_amount" db:"is_incomplete_amount"`
	IsIncompleteSplit  bool     `json:"is_incomplete_split" db:"is_incomplete_split"`
	Latitude           *float64 `json:"latitude,omitempty" db:"latitude"`   // pointer because nullable in db
	Longitude          *float64 `json:"longitude,omitempty" db:"longitude"` // pointer because nullable in db
}

// ExpenseDetails represents detailed information about an expense including its splits
type ExpenseDetails struct {
	Expense                // Struct embedding to include all Expense fields
	Splits  []ExpenseSplit `json:"splits"`
}

// ExpenseSplit represents how an expense is split among users
type ExpenseSplit struct {
	ExpenseID string  `json:"-" db:"expense_id"`
	UserID    string  `json:"user_id" db:"user_id"`
	Amount    float64 `json:"amount" db:"amount"`
	IsPaid    bool    `json:"is_paid" db:"is_paid"` // "paid" or "owes"
}

// Settlement represents the balance between two users (positive = owed, negative = owes)
type Settlement struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"` // positive: user owes you, negative: you owe user
}

// UserSpendings represents a user's spending summary in a group
type UserSpendings struct {
	TotalPaid   float64                `json:"total_paid"`
	TotalOwed   float64                `json:"total_owed"`
	NetSpending float64                `json:"net_spending"`
	Expenses    []UserSpendingsExpense `json:"expenses"`
}

// UserSpendingsExpense extends Expense with user-specific amount
type UserSpendingsExpense struct {
	Expense
	UserAmount float64 `json:"user_amount"` // Amount user paid/owes for this expense
}
