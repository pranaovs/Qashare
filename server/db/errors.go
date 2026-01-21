package db

import (
	"errors"
	"fmt"
	"strings"
)

// Common database errors that can be checked with errors.Is()
var (
	// ErrNotFound indicates that the requested resource was not found in the database
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists indicates that a resource with the same identifier already exists
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput indicates that the provided input is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrPermissionDenied indicates that the user doesn't have permission for the operation
	ErrPermissionDenied = errors.New("permission denied")

	// ErrConflict indicates a conflict with the current state of the resource
	ErrConflict = errors.New("resource conflict")
)

// User-specific errors
var (
	// ErrUserNotFound indicates that the requested user was not found
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailNotRegistered indicates that the email is not registered
	ErrEmailNotRegistered = errors.New("email not registered")

	// ErrEmailAlreadyExists indicates that a user with the email already exists
	ErrEmailAlreadyExists = errors.New("user with this email already exists")

	// ErrUsersNotRelated indicates that users don't share any groups
	ErrUsersNotRelated = errors.New("users not related")
)

// Group-specific errors
var (
	// ErrGroupNotFound indicates that the requested group was not found
	ErrGroupNotFound = errors.New("group not found")

	// ErrNotMember indicates that the user is not a member of the group
	ErrNotMember = errors.New("not a member")

	// ErrNotGroupCreator indicates that the user is not the creator of the group
	ErrNotGroupCreator = errors.New("not group creator")
)

// Expense-specific errors
var (
	// ErrExpenseNotFound indicates that the requested expense was not found
	ErrExpenseNotFound = errors.New("expense not found")

	// ErrInvalidAmount indicates that the expense amount is invalid
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrTitleRequired indicates that the expense title is required
	ErrTitleRequired = errors.New("title required")

	// ErrExpenseIDRequired indicates that the expense ID is required
	ErrExpenseIDRequired = errors.New("expense_id required")
)

// DBError wraps a database error with additional context
type DBError struct {
	Op      string // Operation that failed (e.g., "CreateUser", "GetGroup")
	Err     error  // Underlying error
	Message string // Additional context message
}

func (e *DBError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *DBError) Unwrap() error {
	return e.Err
}

// NewDBError creates a new DBError
func NewDBError(op string, err error, message string) *DBError {
	return &DBError{
		Op:      op,
		Err:     err,
		Message: message,
	}
}

// IsNoRows checks if an error is a "no rows" error
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	// Check for pgx.ErrNoRows (imported in files that use it)
	return err.Error() == "no rows in result set"
}

// IsConstraintViolation checks if an error is a database constraint violation
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL constraint violation errors typically contain "constraint"
	errStr := err.Error()
	return strings.Contains(errStr, "constraint") || strings.Contains(errStr, "violates")
}

// IsDuplicateKey checks if an error is a duplicate key violation
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint")
}
