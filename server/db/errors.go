package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ErrorCode represents a database error category
type ErrorCode string

const (
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeDuplicateKey        ErrorCode = "DUPLICATE_KEY"
	ErrCodeConstraintViolation ErrorCode = "CONSTRAINT_VIOLATION"
	ErrCodeInvalidInput        ErrorCode = "INVALID_INPUT"
)

// DBError represents a database-specific error without HTTP concerns
type DBError struct {
	Code    ErrorCode
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

// WithMessage creates a copy of the error with a custom message
func (e *DBError) WithMessage(format string, args ...interface{}) *DBError {
	newErr := *e
	newErr.Message = fmt.Sprintf(format, args...)
	return &newErr
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
		Code:    ErrCodeNotFound,
		Message: "record not found",
	}

	// ErrDuplicateKey indicates a unique constraint violation
	ErrDuplicateKey = &DBError{
		Code:    ErrCodeDuplicateKey,
		Message: "duplicate key violation",
	}

	// ErrConstraintViolation indicates a database constraint was violated
	ErrConstraintViolation = &DBError{
		Code:    ErrCodeConstraintViolation,
		Message: "constraint violation",
	}

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = &DBError{
		Code:    ErrCodeInvalidInput,
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
		return dbErr.Code == ErrCodeNotFound
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
		return dbErr.Code == ErrCodeDuplicateKey
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
		return dbErr.Code == ErrCodeConstraintViolation
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
