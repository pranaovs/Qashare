package apierrors

import (
	"fmt"
)

type AppError struct {
	HTTPCode    int    `json:"-"`       // e.g., 400, 404
	MachineCode string `json:"code"`    // e.g., "BAD_NAME", "INVALID_EMAIL"
	Message     string `json:"message"` // Human-readable message
	Err         error  `json:"-"`       // Internal error for logging (optional)
}

// WithInternal creates a COPY of the error and attaches the internal error.
// It returns a pointer so you can return it directly.
func (e *AppError) WithInternal(err error) *AppError {
	// Create a shallow copy of the struct value.
	// This ensures we don't modify the global variable (like ErrInvalidEmail).
	newErr := *e

	// Attach the internal error to the copy.
	newErr.Err = err

	return &newErr
}

// Error implements the standard error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap allows standard Go error wrapping/unwrapping.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError.
func New(httpCode int, machineCode string, message string, err error) *AppError {
	return &AppError{
		HTTPCode:    httpCode,
		MachineCode: machineCode,
		Message:     message,
		Err:         err,
	}
}

// Msg creates a clone of the error with a new custom message.
// It preserves the HTTPCode and MachineCode from the original error.
func (e *AppError) Msg(msg string) *AppError {
	newErr := *e
	newErr.Message = msg
	return &newErr
}

// Msgf creates a clone of the error with a formatted custom message.
// It preserves the HTTPCode and MachineCode from the original error.
func (e *AppError) Msgf(format string, args ...any) *AppError {
	newErr := *e
	newErr.Message = fmt.Sprintf(format, args...)
	return &newErr
}
