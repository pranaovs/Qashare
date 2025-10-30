package utils

// GetUniqueUserIDs extracts unique user IDs from a slice of user IDs.
// This handles cases where the same user appears multiple times in splits
// (e.g., once as is_paid=true and once as is_paid=false).
func GetUniqueUserIDs(userIDs []string) []string {
	seen := make(map[string]bool)
	unique := make([]string, 0, len(userIDs))

	for _, id := range userIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
}
