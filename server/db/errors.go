package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// DBError represents a database-specific error without HTTP concerns.
// It implements the apperrors.AppError interface for consistent error handling.
type DBError struct {
	Code    string
	Message string
	Err     error // underlying error for debugging
}

// Error implements the error interface
func (e *DBError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (underlying: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap allows error unwrapping for errors.Is and errors.As
func (e *DBError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is matching - allows errors.Is to match derived errors
func (e *DBError) Is(target error) bool {
	t, ok := target.(*DBError)
	if !ok {
		return false
	}
	// Match if codes are equal
	return e.Code == t.Code
}

// Msg creates a copy of the error with a custom message
func (e *DBError) Msg(msg string) *DBError {
	newErr := *e
	newErr.Message = msg
	return &newErr
}

// Msgf creates a copy of the error with a formatted custom message
func (e *DBError) Msgf(format string, args ...any) *DBError {
	newErr := *e
	newErr.Message = fmt.Sprintf(format, args...)
	return &newErr
}

// GetCode returns the error code (implements DomainError interface)
func (e *DBError) GetCode() string {
	return e.Code
}

// GetMessage returns the error message (implements DomainError interface)
func (e *DBError) GetMessage() string {
	return e.Message
}

// WithError creates a copy of the error and attaches an underlying error
func (e *DBError) WithError(err error) *DBError {
	newErr := *e
	newErr.Err = err
	return &newErr
}

// Sentinel errors for common database operations
var (
	// ErrNotFound indicates a record was not found
	ErrNotFound = &DBError{
		Code:    "NOT_FOUND",
		Message: "record not found",
	}

	// ErrDuplicateKey indicates a unique constraint violation
	ErrDuplicateKey = &DBError{
		Code:    "DUPLICATE_KEY",
		Message: "duplicate key violation",
	}

	// ErrConstraintViolation indicates a database constraint was violated
	ErrConstraintViolation = &DBError{
		Code:    "CONSTRAINT_VIOLATION",
		Message: "constraint violation",
	}

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = &DBError{
		Code:    "INVALID_INPUT",
		Message: "invalid input data",
	}
)

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var dbErr *DBError
	if errors.As(err, &dbErr) {
		return dbErr.Code == ErrNotFound.Code
	}
	return false
}

// IsDuplicate checks if an error is a duplicate key error
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	var dbErr *DBError
	if errors.As(err, &dbErr) {
		return dbErr.Code == ErrDuplicateKey.Code
	}
	return false
}

// IsConstraintError checks if an error is a constraint violation
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}
	var dbErr *DBError
	if errors.As(err, &dbErr) {
		return dbErr.Code == ErrConstraintViolation.Code
	}
	return false
}

// IsNoRows checks if an error is a "no rows" error from pgx
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	return err == pgx.ErrNoRows || err.Error() == "no rows in result set"
}

// IsConstraintViolation checks if an error is a database constraint violation
// This checks the raw database error, not our wrapped DBError
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "constraint") || strings.Contains(errStr, "violates")
}

// IsDuplicateKey checks if an error is a duplicate key violation
// This checks the raw database error, not our wrapped DBError
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint")
}
