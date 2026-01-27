package handlers

import (
	"errors"
	"github.com/pranaovs/qashare/db"
	"github.com/pranaovs/qashare/models"
)

// mapDBError maps database errors to structured API error responses
func mapDBError(err error) models.ErrorResponse {
	if err == nil {
		return models.NewSimpleErrorResponse("unknown error", models.ErrCodeInternal)
	}

	// Check for specific database errors
	switch {
	case errors.Is(err, db.ErrUserNotFound):
		return models.NewSimpleErrorResponse("user not found", models.ErrCodeUserNotFound)
	case errors.Is(err, db.ErrEmailNotRegistered):
		return models.NewSimpleErrorResponse("email not registered", models.ErrCodeEmailNotRegistered)
	case errors.Is(err, db.ErrEmailAlreadyExists):
		return models.NewSimpleErrorResponse("user with this email already exists", models.ErrCodeEmailExists)
	case errors.Is(err, db.ErrUsersNotRelated):
		return models.NewSimpleErrorResponse("users not related", models.ErrCodeUsersNotRelated)
	case errors.Is(err, db.ErrGroupNotFound):
		return models.NewSimpleErrorResponse("group not found", models.ErrCodeGroupNotFound)
	case errors.Is(err, db.ErrNotMember):
		return models.NewSimpleErrorResponse("not a member", models.ErrCodeNotMember)
	case errors.Is(err, db.ErrNotGroupCreator):
		return models.NewSimpleErrorResponse("not group creator", models.ErrCodeNotGroupCreator)
	case errors.Is(err, db.ErrExpenseNotFound):
		return models.NewSimpleErrorResponse("expense not found", models.ErrCodeExpenseNotFound)
	case errors.Is(err, db.ErrInvalidAmount):
		return models.NewSimpleErrorResponse("invalid amount", models.ErrCodeInvalidAmount)
	case errors.Is(err, db.ErrTitleRequired):
		return models.NewSimpleErrorResponse("title required", models.ErrCodeTitleRequired)
	case errors.Is(err, db.ErrExpenseIDRequired):
		return models.NewSimpleErrorResponse("expense_id required", models.ErrCodeExpenseIDRequired)
	case errors.Is(err, db.ErrNotFound):
		return models.NewSimpleErrorResponse("resource not found", models.ErrCodeNotFound)
	case errors.Is(err, db.ErrAlreadyExists):
		return models.NewSimpleErrorResponse("resource already exists", models.ErrCodeConflict)
	case errors.Is(err, db.ErrInvalidInput):
		return models.NewSimpleErrorResponse("invalid input", models.ErrCodeInvalidInput)
	case errors.Is(err, db.ErrPermissionDenied):
		return models.NewSimpleErrorResponse("permission denied", models.ErrCodeForbidden)
	case errors.Is(err, db.ErrConflict):
		return models.NewSimpleErrorResponse("resource conflict", models.ErrCodeConflict)
	default:
		// For unknown errors, return internal error
		return models.NewErrorResponse("internal server error", models.ErrCodeInternal, err.Error())
	}
}
