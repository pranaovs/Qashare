package apperrors

import "net/http"

// Define your reusable "Templates" for errors here
var (
	ErrInvalidName  = New(http.StatusBadRequest, "BAD_NAME", "The name provided contains invalid characters.", nil)
	ErrInvalidEmail = New(http.StatusBadRequest, "BAD_EMAIL", "The email format is incorrect.", nil)
	ErrUserNotFound = New(http.StatusNotFound, "USER_NOT_FOUND", "The requested user does not exist.", nil)
)
