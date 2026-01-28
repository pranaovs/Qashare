package apperrors

import (
	"errors"

	"github.com/pranaovs/qashare/routes/apierrors"
)

func MapError(err error, errorMap map[error]*apierrors.AppError) error {
	if err == nil {
		return nil
	}

	for specificErr, apiErr := range errorMap {
		if errors.Is(err, specificErr) {
			// Check if the error implements AppError interface
			var appErr AppError
			if errors.As(err, &appErr) {
				// Get the default message from the sentinel error
				var defaultErr AppError
				if errors.As(specificErr, &defaultErr) {
					// Compare messages - if different, it means custom message was set
					if appErr.GetMessage() != defaultErr.GetMessage() {
						// Custom message provided - propagate it to API error
						return apiErr.Msg(appErr.GetMessage())
					}
				}
			}

			// No custom message - return default API error
			return apiErr
		}
	}

	// Unknown error - return err itself
	return err
}
