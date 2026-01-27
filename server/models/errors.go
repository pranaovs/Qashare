package models

// ErrorResponse represents a standardized error response for the API
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid input"`
	Code    string `json:"code" example:"Bad Request"`
	Message string `json:"message,omitempty" example:"The provided email format is invalid"`
}

// ErrBadRequest represents a 400 Bad Request validation error
type ErrBadRequest struct {
	Error   string `json:"error" example:"invalid request body"`
	Code    string `json:"code" example:"Bad Request"`
	Message string `json:"message,omitempty" example:"The provided email format is invalid"`
}

// ErrUnauthorized represents a 401 Unauthorized error
type ErrUnauthorized struct {
	Error   string `json:"error" example:"invalid token"`
	Code    string `json:"code" example:"Unauthorized"`
	Message string `json:"message,omitempty" example:""`
}

// ErrForbidden represents a 403 Forbidden error
type ErrForbidden struct {
	Error   string `json:"error" example:"access denied"`
	Code    string `json:"code" example:"Forbidden"`
	Message string `json:"message,omitempty" example:""`
}

// ErrNotFound represents a 404 Not Found error
type ErrNotFound struct {
	Error   string `json:"error" example:"resource not found"`
	Code    string `json:"code" example:"Not Found"`
	Message string `json:"message,omitempty" example:""`
}

// ErrConflict represents a 409 Conflict error
type ErrConflict struct {
	Error   string `json:"error" example:"resource already exists"`
	Code    string `json:"code" example:"Conflict"`
	Message string `json:"message,omitempty" example:""`
}

// ErrInternalServer represents a 500 Internal Server Error
type ErrInternalServer struct {
	Error   string `json:"error" example:"internal server error"`
	Code    string `json:"code" example:"Internal Server Error"`
	Message string `json:"message,omitempty" example:"An unexpected error occurred"`
}

// ErrorCode constants for stable error identification using standard HTTP status text
const (
	// General errors - using standard HTTP status text
	ErrCodeInternal     = "Internal Server Error"
	ErrCodeBadRequest   = "Bad Request"
	ErrCodeValidation   = "Bad Request"
	ErrCodeNotFound     = "Not Found"
	ErrCodeConflict     = "Conflict"
	ErrCodeUnauthorized = "Unauthorized"
	ErrCodeForbidden    = "Forbidden"
	ErrCodeInvalidInput = "Bad Request"

	// User-specific errors
	ErrCodeUserNotFound       = "Not Found"
	ErrCodeEmailExists        = "Conflict"
	ErrCodeEmailNotRegistered = "Not Found"
	ErrCodeUsersNotRelated    = "Forbidden"
	ErrCodeInvalidCredentials = "Unauthorized"
	ErrCodeInvalidToken       = "Unauthorized"
	ErrCodeExpiredToken       = "Unauthorized"

	// Group-specific errors
	ErrCodeGroupNotFound   = "Not Found"
	ErrCodeNotMember       = "Forbidden"
	ErrCodeNotGroupCreator = "Forbidden"

	// Expense-specific errors
	ErrCodeExpenseNotFound   = "Not Found"
	ErrCodeInvalidAmount     = "Bad Request"
	ErrCodeTitleRequired     = "Bad Request"
	ErrCodeExpenseIDRequired = "Bad Request"
)

// NewErrorResponse creates a new error response with code and message
func NewErrorResponse(error, code, message string) ErrorResponse {
	return ErrorResponse{
		Error:   error,
		Code:    code,
		Message: message,
	}
}

// NewSimpleErrorResponse creates a new error response with just error and code
func NewSimpleErrorResponse(error, code string) ErrorResponse {
	return ErrorResponse{
		Error: error,
		Code:  code,
	}
}
