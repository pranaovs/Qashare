package db

import (
	"strings"
)

// IsNoRows checks if an error is a "no rows" error
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	// Check for pgx.ErrNoRows (imported in files that use it)
	return err.Error() == "no rows in result set"
}

// IsConstraintViolation checks if an error is a database constraint violation
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL constraint violation errors typically contain "constraint"
	errStr := err.Error()
	return strings.Contains(errStr, "constraint") || strings.Contains(errStr, "violates")
}

// IsDuplicateKey checks if an error is a duplicate key violation
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint")
}
