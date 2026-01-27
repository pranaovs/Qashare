package utils

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pranaovs/qashare/apierrors"
)

func SendError(c *gin.Context, err error) {
	// Check if the error is our custom AppError
	if appErr, ok := err.(*apierrors.AppError); ok {

		LogDebug(c, fmt.Sprintf("Error: %s | Code: %s | Internal: %v",
			appErr.Message, appErr.MachineCode, appErr.Err))

		// Send the encapsulated response and return
		c.JSON(appErr.HTTPCode, gin.H{
			"code":    appErr.MachineCode,
			"message": appErr.Message,
		})
		return
	}

	// Handle unexpected/unknown errors (Panic recovery or generic errors)
	fmt.Fprintf(os.Stderr, "[ERROR] Internal Server Error: %v\n", err)

	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_ERROR",
		"message": "Something went wrong on our end.",
	})
}
