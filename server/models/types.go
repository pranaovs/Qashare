// Package models defines the core data structures for the shared expenses application.
package models

import "github.com/google/uuid"

// User represents a user in the system
type User struct {
	UserID       uuid.UUID `json:"user_id" db:"user_id" immutable:"true"`
	Name         string    `json:"name" db:"user_name"`
	Email        string    `json:"email" db:"email"`
	Guest        bool      `json:"guest" db:"is_guest" immutable:"true"`
	PasswordHash *string   `json:"-" db:"password_hash" immutable:"true"` // excluded from JSON responses
	CreatedAt    int64     `json:"created_at" db:"created_at" immutable:"true"`
}

// Group represents a group
type Group struct {
	GroupID     uuid.UUID `json:"group_id" db:"group_id" immutable:"true"`
	Name        string    `json:"name" db:"group_name"`
	Description string    `json:"description" db:"description"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by" immutable:"true"`
	CreatedAt   int64     `json:"created_at" db:"created_at" immutable:"true"`
}

// GroupDetails represents detailed information about a group including its members
type GroupDetails struct {
	Group               // Struct embedding to include all Group fields
	Members []GroupUser `json:"members"`
}

// GroupMember represents a user's membership in a group
type GroupMember struct {
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	GroupID  uuid.UUID `json:"group_id" db:"group_id"`
	JoinedAt int64     `json:"joined_at" db:"joined_at"`
}

// GroupUser Not a part of DB schema, used for responses
type GroupUser struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Guest    bool      `json:"guest"`
	JoinedAt int64     `json:"joined_at"`
}

// Expense represents an expense in a group(ID)
type Expense struct {
	ExpenseID          uuid.UUID `json:"expense_id" db:"expense_id" immutable:"true"`
	GroupID            uuid.UUID `json:"group_id" db:"group_id" immutable:"true"`
	AddedBy            uuid.UUID `json:"added_by" db:"added_by" immutable:"true"`
	Title              string    `json:"title" db:"title"`
	Description        *string   `json:"description" db:"description"` // pointer because nullable in db
	CreatedAt          int64     `json:"created_at" db:"created_at" immutable:"true"`
	TransactedAt       *int64    `json:"transacted_at" db:"transacted_at"`
	Amount             float64   `json:"amount" db:"amount"`
	IsIncompleteAmount bool      `json:"is_incomplete_amount" db:"is_incomplete_amount"`
	IsIncompleteSplit  bool      `json:"is_incomplete_split" db:"is_incomplete_split"`
	IsSettlement       bool      `json:"is_settlement" db:"is_settlement" immutable:"true"`
	Latitude           *float64  `json:"latitude" db:"latitude"`   // pointer because nullable in db
	Longitude          *float64  `json:"longitude" db:"longitude"` // pointer because nullable in db
}

// ExpenseDetails represents detailed information about an expense including its splits
type ExpenseDetails struct {
	Expense                // Struct embedding to include all Expense fields
	Splits  []ExpenseSplit `json:"splits"`
}

// ExpenseSplit represents how an expense is split among users
type ExpenseSplit struct {
	ExpenseID uuid.UUID `json:"-" db:"expense_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Amount    float64   `json:"amount" db:"amount"`
	IsPaid    bool      `json:"is_paid" db:"is_paid"` // "paid" or "owes"
}

// Settlement represents a balance or transaction between two users, used for responses.
// Settlement data is stored as an Expense with IsSettlement=true in the DB.
//
// Amount sign is relative to the authenticated user:
//   - Positive: you are owed by / paid UserID (net credit)
//   - Negative: you owe / were paid by UserID (net debt)
//
// In the GetSettle endpoint (balance computation), this is the net amount.
// In settlement history and CRUD, this reflects the payment direction.
type Settlement struct {
	Title        string    `json:"title"`
	CreatedAt    int64     `json:"created_at" immutable:"true"`
	GroupID      uuid.UUID `json:"group_id" immutable:"true"`
	TransactedAt *int64    `json:"transacted_at"`
	UserID       uuid.UUID `json:"user_id" immutable:"true"` // The other user involved in the settlement
	Amount       float64   `json:"amount"`
}

// UserExpense extends Expense with user-specific amount
type UserExpense struct {
	Expense
	UserAmount float64 `json:"user_amount"` // Amount user paid/owes for this expense
}
