package apperrors

import (
	"fmt"
)

type AppError struct {
	HTTPCode    int    `json:"-"`       // e.g., 400, 404
	MachineCode string `json:"code"`    // e.g., "BAD_NAME", "INVALID_EMAIL"
	Message     string `json:"message"` // Human-readable message
	Err         error  `json:"-"`       // Internal error for logging (optional)
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
