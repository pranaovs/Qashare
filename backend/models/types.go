// Package models defines the core data structures for the shared expenses application.
package models

type User struct {
	UserID       string  `json:"user_id" db:"user_id"`
	Name         string  `json:"name" db:"user_name"`
	Email        string  `json:"email" db:"email"`
	Guest        bool    `json:"guest" db:"is_guest"`
	PasswordHash *string `json:"-" db:"password_hash"` // excluded from JSON responses
	CreatedAt    int64   `json:"created_at" db:"created_at"`
}

type Group struct {
	GroupID     string `json:"group_id" db:"group_id"`
	Name        string `json:"name" db:"group_name"`
	Description string `json:"description,omitempty" db:"description"`
	CreatedBy   string `json:"created_by" db:"created_by"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`

	Members []GroupUser `json:"members" db:"-"` // NOTE: Be careful with this, not a part of DB schema
}

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

type Expense struct {
	ExpenseID          string  `json:"expense_id" db:"expense_id"`
	GroupID            string  `json:"group_id" db:"group_id"`
	AddedBy            string  `json:"added_by" db:"added_by"`
	Title              string  `json:"title" db:"title"`
	Description        string  `json:"description,omitempty" db:"description"`
	CreatedAt          int64   `json:"created_at" db:"created_at"`
	Amount             float64 `json:"amount" db:"amount"`
	IsIncompleteAmount bool    `json:"is_incomplete_amount" db:"is_incomplete_amount"`
	IsIncompleteSplit  bool    `json:"is_incomplete_split" db:"is_incomplete_split"`
	Latitude           float64 `json:"latitude,omitempty" db:"latitude"`
	Longitude          float64 `json:"longitude,omitempty" db:"longitude"`

	Splits []ExpenseSplit `json:"splits" db:"-"`
}

type ExpenseSplit struct {
	ExpenseID string  `json:"-" db:"expense_id"`
	UserID    string  `json:"user_id" db:"user_id"`
	Amount    float64 `json:"amount" db:"amount"`
	IsPaid    bool    `json:"is_paid" db:"is_paid"` // "paid" or "owes"
}
