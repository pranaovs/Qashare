package models

// ErrorResponse represents a standardized error response for the API
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid input"`
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message,omitempty" example:"The provided email format is invalid"`
}

// ErrorCode constants for stable error identification
const (
	// General errors
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeInvalidInput   = "INVALID_INPUT"

	// User-specific errors
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeEmailExists        = "EMAIL_ALREADY_EXISTS"
	ErrCodeEmailNotRegistered = "EMAIL_NOT_REGISTERED"
	ErrCodeUsersNotRelated    = "USERS_NOT_RELATED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeExpiredToken       = "EXPIRED_TOKEN"

	// Group-specific errors
	ErrCodeGroupNotFound    = "GROUP_NOT_FOUND"
	ErrCodeNotMember        = "NOT_MEMBER"
	ErrCodeNotGroupCreator  = "NOT_GROUP_CREATOR"

	// Expense-specific errors
	ErrCodeExpenseNotFound     = "EXPENSE_NOT_FOUND"
	ErrCodeInvalidAmount       = "INVALID_AMOUNT"
	ErrCodeTitleRequired       = "TITLE_REQUIRED"
	ErrCodeExpenseIDRequired   = "EXPENSE_ID_REQUIRED"
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
