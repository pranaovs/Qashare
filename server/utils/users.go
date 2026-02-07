package utils

import "github.com/google/uuid"

// GetUniqueUserIDs extracts unique user IDs from a slice of user IDs.
// This handles cases where the same user appears multiple times in splits
// (e.g., once as is_paid=true and once as is_paid=false).
func GetUniqueUserIDs(userIDs []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	unique := make([]uuid.UUID, 0, len(userIDs))

	for _, id := range userIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
}
