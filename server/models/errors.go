package models

// ErrorResponse represents a standardized error response for the API
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid input"`
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message,omitempty" example:"The provided email format is invalid"`
}

// ValidationErrorResponse represents a 400 Bad Request validation error
type ValidationErrorResponse struct {
	Error   string `json:"error" example:"invalid request body"`
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message,omitempty" example:"The provided email format is invalid"`
}

// UnauthorizedErrorResponse represents a 401 Unauthorized error
type UnauthorizedErrorResponse struct {
	Error   string `json:"error" example:"invalid token"`
	Code    string `json:"code" example:"INVALID_TOKEN"`
	Message string `json:"message,omitempty" example:""`
}

// ForbiddenErrorResponse represents a 403 Forbidden error
type ForbiddenErrorResponse struct {
	Error   string `json:"error" example:"access denied"`
	Code    string `json:"code" example:"FORBIDDEN"`
	Message string `json:"message,omitempty" example:""`
}

// NotFoundErrorResponse represents a 404 Not Found error
type NotFoundErrorResponse struct {
	Error   string `json:"error" example:"resource not found"`
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message,omitempty" example:""`
}

// ConflictErrorResponse represents a 409 Conflict error
type ConflictErrorResponse struct {
	Error   string `json:"error" example:"resource already exists"`
	Code    string `json:"code" example:"CONFLICT"`
	Message string `json:"message,omitempty" example:""`
}

// InternalErrorResponse represents a 500 Internal Server Error
type InternalErrorResponse struct {
	Error   string `json:"error" example:"internal server error"`
	Code    string `json:"code" example:"INTERNAL_ERROR"`
	Message string `json:"message,omitempty" example:"An unexpected error occurred"`
}

// ErrorCode constants for stable error identification
const (
	// General errors
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeInvalidInput = "INVALID_INPUT"

	// User-specific errors
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeEmailExists        = "EMAIL_ALREADY_EXISTS"
	ErrCodeEmailNotRegistered = "EMAIL_NOT_REGISTERED"
	ErrCodeUsersNotRelated    = "USERS_NOT_RELATED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeExpiredToken       = "EXPIRED_TOKEN"

	// Group-specific errors
	ErrCodeGroupNotFound   = "GROUP_NOT_FOUND"
	ErrCodeNotMember       = "NOT_MEMBER"
	ErrCodeNotGroupCreator = "NOT_GROUP_CREATOR"

	// Expense-specific errors
	ErrCodeExpenseNotFound   = "EXPENSE_NOT_FOUND"
	ErrCodeInvalidAmount     = "INVALID_AMOUNT"
	ErrCodeTitleRequired     = "TITLE_REQUIRED"
	ErrCodeExpenseIDRequired = "EXPENSE_ID_REQUIRED"
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
