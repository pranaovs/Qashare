package apperrors

// AppError is a common interface for all app-specific errors
// (e.g., db.DBError, utils.UtilsError). It allows generic error handling
// without explicit type dependencies.
//
// Both DBError and UtilsError implement this interface, enabling:
//   - Type-safe error handling across layers
//   - Custom message propagation through MapError
//   - Consistent error code and message retrieval
type AppError interface {
	error
	GetCode() string
	GetMessage() string
}
