package apierrors

import "net/http"

var (
	ErrBadRequest         = New(http.StatusBadRequest, "BAD_REQUEST", "The request is invalid or malformed.", nil)
	ErrInvalidName        = New(http.StatusBadRequest, "BAD_NAME", "The name provided contains invalid characters.", nil)
	ErrEmailAlreadyExists = New(http.StatusConflict, "EMAIL_EXISTS", "An account with this email already exists.", nil)
	ErrInvalidEmail       = New(http.StatusBadRequest, "BAD_EMAIL", "The email format is incorrect.", nil)
	ErrInvalidDescription = New(http.StatusBadRequest, "BAD_DESCRIPTION", "The description contains invalid characters.", nil)

	// Auth Errors
	ErrInvalidPassword               = New(http.StatusBadRequest, "BAD_PASSWORD", "The password syntax is incorrect.", nil)
	ErrBadCredentials                = New(http.StatusUnauthorized, "BAD_CREDENTIALS", "The provided credentials are incorrect.", nil)
	ErrInvalidAccessToken            = New(http.StatusUnauthorized, "INVALID_TOKEN", "The access token is invalid.", nil)
	ErrExpiredAccessToken            = New(http.StatusForbidden, "EXPIRED_TOKEN", "The access token has expired.", nil)
	ErrInvalidRefreshToken           = New(http.StatusBadRequest, "INVALID_REFRESH_TOKEN", "The refresh token is invalid.", nil)
	ErrExpiredRefreshToken           = New(http.StatusForbidden, "EXPIRED_REFRESH_TOKEN", "The refresh token has expired.", nil)
	ErrEmailNotVerified              = New(http.StatusForbidden, "EMAIL_NOT_VERIFIED", "The email address has not been verified.", nil)
	ErrEmailVerificationTokenExpired = New(http.StatusForbidden, "EMAIL_VERIFICATION_TOKEN_EXPIRED", "The email verification token has expired.", nil)
	ErrEmailVerificationTokenError   = New(http.StatusBadRequest, "EMAIL_VERIFICATION_TOKEN_ERROR", "The email verification token is invalid or malformed.", nil)

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
