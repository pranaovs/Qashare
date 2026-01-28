package utils

import (
	"fmt"
)

// UtilsError represents a utils-specific error without HTTP concerns.
// It implements the apperrors.AppError interface for consistent error handling.
type UtilsError struct {
	Code    string
	Message string
	Err     error // underlying error for debugging
}

// Error implements the error interface
func (e *UtilsError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (underlying: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap allows error unwrapping for errors.Is and errors.As
func (e *UtilsError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is matching - allows errors.Is to match derived errors
func (e *UtilsError) Is(target error) bool {
	t, ok := target.(*UtilsError)
	if !ok {
		return false
	}
	// Match if codes are equal
	return e.Code == t.Code
}

// Msg creates a copy of the error with a custom message
func (e *UtilsError) Msg(msg string) *UtilsError {
	newErr := *e
	newErr.Message = msg
	return &newErr
}

// Msgf creates a copy of the error with a formatted custom message
func (e *UtilsError) Msgf(format string, args ...any) *UtilsError {
	newErr := *e
	newErr.Message = fmt.Sprintf(format, args...)
	return &newErr
}

// WithError creates a copy of the error and attaches an underlying error
func (e *UtilsError) WithError(err error) *UtilsError {
	newErr := *e
	newErr.Err = err
	return &newErr
}

// GetCode returns the error code (implements DomainError interface)
func (e *UtilsError) GetCode() string {
	return e.Code
}

// GetMessage returns the error message (implements DomainError interface)
func (e *UtilsError) GetMessage() string {
	return e.Message
}

// Sentinel errors for common validation operations
var (
	// ErrInvalidName indicates an invalid name format
	ErrInvalidName = &UtilsError{
		Code:    "INVALID_NAME",
		Message: "invalid name format",
	}

	// ErrInvalidEmail indicates an invalid email format
	ErrInvalidEmail = &UtilsError{
		Code:    "INVALID_EMAIL",
		Message: "invalid email format",
	}

	// ErrInvalidPassword indicates an invalid password
	ErrInvalidPassword = &UtilsError{
		Code:    "INVALID_PASSWORD",
		Message: "invalid password",
	}

	// ErrHashingFailed indicates password hashing failure
	ErrHashingFailed = &UtilsError{
		Code:    "HASHING_FAILED",
		Message: "failed to hash password",
	}

	// ErrInvalidToken indicates an invalid token
	ErrInvalidToken = &UtilsError{
		Code:    "INVALID_TOKEN",
		Message: "invalid token",
	}
)
