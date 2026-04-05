package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// parseCursor extracts and parses the "cursor" query parameter as a UUID.
// Returns nil if the parameter is absent (first page request).
func parseCursor(c *gin.Context) (*uuid.UUID, error) {
	cursorStr := c.Query("cursor")
	if cursorStr == "" {
		return nil, nil
	}
	id, err := uuid.Parse(cursorStr)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
