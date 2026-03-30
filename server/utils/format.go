package utils

import (
	"fmt"
	"math"
	"time"
)

// FormatDuration returns a human-readable string like "24 hours" or "2 hours 30 minutes".
func FormatDuration(d time.Duration) string {
	h := int(math.Floor(d.Hours()))
	m := int(d.Minutes()) % 60

	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%d hours %d minutes", h, m)
	case h > 0:
		return fmt.Sprintf("%d hours", h)
	case m > 0:
		return fmt.Sprintf("%d minutes", m)
	default:
		return "a few seconds"
	}
}
