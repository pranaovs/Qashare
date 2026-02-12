package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pranaovs/qashare/routes/apierrors"
)

// SendError inspects the provided error and sends an appropriate JSON response.
// This function differentiates between known application errors and unexpected errors.
// Application errors are sent with their specific HTTP status codes and messages,
// Generic errors result in a 500 Internal Server Error response.
func SendError(c *gin.Context, err error) {
	// Check if the error is our custom AppError
	if appErr, ok := err.(*apierrors.AppError); ok {

		LogDebug(c.Request.Context(), fmt.Sprintf("Error: %s | Code: %s | Internal: %v",
			appErr.Message, appErr.MachineCode, appErr.Err))

		// Send the encapsulated response and return
		c.JSON(appErr.HTTPCode, gin.H{
			"code":    appErr.MachineCode,
			"message": appErr.Message,
		})
		return
	}

	// Handle unexpected/unknown errors (Panic recovery or generic errors)
	LogError(c.Request.Context(), "internal server error", err)

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_ERROR",
		"message": "Something went wrong on our end. Please report this.",
	})
}

// SendAbort aborts the request and sends a JSON error response using the same
// {"code", "message"} format as SendError for consistent error responses.
func SendAbort(c *gin.Context, appErr *apierrors.AppError) {
	c.AbortWithStatusJSON(appErr.HTTPCode, gin.H{
		"code":    appErr.MachineCode,
		"message": appErr.Message,
	})
}

// SendJSON is a helper function that sends a JSON response with the specified
// HTTP status code and data.
func SendJSON(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, data)
}

// SendOK sends a standard OK response with a message.
func SendOK(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// SendData sends a standard OK response with arbitrary data.
func SendData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}
