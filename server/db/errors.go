package db

import (
	"strings"

	"github.com/jackc/pgx/v5"
)

// IsNoRows checks if an error is a "no rows" error from pgx
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	return err == pgx.ErrNoRows || err.Error() == "no rows in result set"
}

// IsConstraintViolation checks if an error is a database constraint violation
// This checks the raw database error, not our wrapped DBError
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "constraint") || strings.Contains(errStr, "violates")
}

// IsDuplicateKey checks if an error is a duplicate key violation
// This checks the raw database error, not our wrapped DBError
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint")
}
