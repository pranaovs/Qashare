package apierrors

import "net/http"

var (
	ErrBadRequest         = New(http.StatusBadRequest, "BAD_REQUEST", "The request is invalid or malformed.", nil)
	ErrInvalidName        = New(http.StatusBadRequest, "BAD_NAME", "The name provided contains invalid characters.", nil)
	ErrEmailAlreadyExists = New(http.StatusConflict, "EMAIL_EXISTS", "An account with this email already exists.", nil)
	ErrInvalidEmail       = New(http.StatusBadRequest, "BAD_EMAIL", "The email format is incorrect.", nil)
	ErrInvalidDescription = New(http.StatusBadRequest, "BAD_DESCRIPTION", "The description contains invalid characters.", nil)

	// Auth Errors
	ErrInvalidPassword = New(http.StatusBadRequest, "BAD_PASSWORD", "The password syntax is incorrect.", nil)
	ErrBadCredentials  = New(http.StatusUnauthorized, "BAD_CREDENTIALS", "The provided credentials are incorrect.", nil)
	ErrInvalidToken    = New(http.StatusUnauthorized, "INVALID_TOKEN", "The authentication token is invalid or expired.", nil)

	// Group Errors
	ErrUserNotFound    = New(http.StatusNotFound, "USER_NOT_FOUND", "The requested user does not exist.", nil)
	ErrGroupNotFound   = New(http.StatusNotFound, "GROUP_NOT_FOUND", "The requested group does not exist.", nil)
	ErrUserNotInGroup  = New(http.StatusForbidden, "USER_NOT_IN_GROUP", "The user is not a member of the specified group.", nil)
	ErrUsersNotRelated = New(http.StatusForbidden, "USERS_NOT_RELATED", "The users are not related in the specified context.", nil)
	ErrNoPermissions   = New(http.StatusForbidden, "NO_PERMISSIONS", "You do not have sufficient permissions to perform this action.", nil)
	ErrUserOwnsGroups  = New(http.StatusConflict, "USER_OWNS_GROUPS", "Cannot delete account while owning groups. Transfer ownership first.", nil)

	// Expenses errors
	ErrExpenseNotFound = New(http.StatusNotFound, "EXPENSE_NOT_FOUND", "The requested expense does not exist.", nil)
	ErrInvalidAmount   = New(http.StatusBadRequest, "INVALID_AMOUNT", "The expense amount is invalid.", nil)
	ErrInvalidSplit    = New(http.StatusBadRequest, "INVALID_SPLIT", "The expense splits are invalid or do not sum up correctly.", nil)

	// Generic errors
	ErrInternalServer = New(http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong on our end.", nil)
)
