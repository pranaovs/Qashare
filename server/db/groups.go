// Package db provides database operations for group management.
// This file contains all group-related database operations including creating groups,
// managing group members, and retrieving group information.
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pranaovs/qashare/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateGroup creates a new group in the database and automatically adds the creator as a member.
// This operation is atomic - either both the group creation and membership addition succeed,
// or neither does (using a transaction).
//
// Parameters:
//   - name: The name of the group
//   - description: Optional description of the group
//   - ownerUserID: The ID of the user creating the group
//
// Returns the newly created group's ID or an error if the operation fails.
func CreateGroup(ctx context.Context, pool *pgxpool.Pool, name, description, ownerUserID string) (string, error) {
	log.Printf("[DB] Creating new group: %s (owner: %s)", name, ownerUserID)

	var groupID string

	// Use WithTransaction helper for consistent transaction management
	err := WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Insert the group
		query := `INSERT INTO groups (group_name, description, created_by, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING group_id`

		err := tx.QueryRow(ctx, query, name, description, ownerUserID, time.Now()).Scan(&groupID)
		if err != nil {
			return fmt.Errorf("failed to insert group: %w", err)
		}

		// Add creator as the first member
		memberQuery := `INSERT INTO group_members (user_id, group_id, joined_at)
			VALUES ($1, $2, $3)`

		_, err = tx.Exec(ctx, memberQuery, ownerUserID, groupID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to add creator as member: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", NewDBError("CreateGroup", err, "failed to create group")
	}

	log.Printf("[DB] ✓ Group created successfully with ID: %s", groupID)
	return groupID, nil
}

// GetGroupCreator retrieves the user ID of the group creator.
// This is a lightweight query that only returns the creator ID, useful for authorization checks.
// Returns ErrGroupNotFound if no group with the ID exists.
func GetGroupCreator(ctx context.Context, pool *pgxpool.Pool, groupID string) (string, error) {
	log.Printf("[DB] Fetching group creator for group: %s", groupID)

	var creatorID string
	query := `SELECT created_by FROM groups WHERE group_id = $1`

	err := pool.QueryRow(ctx, query, groupID).Scan(&creatorID)
	if err == pgx.ErrNoRows {
		log.Printf("[DB] Group not found: %s", groupID)
		return "", ErrGroupNotFound
	}
	if err != nil {
		return "", NewDBError("GetGroupCreator", err, "failed to query group creator")
	}

	log.Printf("[DB] ✓ Group creator retrieved: %s", creatorID)
	return creatorID, nil
}

// GetGroup retrieves complete group information including all members.
// Returns a Group struct with full details and a list of all group members.
// Returns ErrGroupNotFound if no group with the ID exists.
func GetGroup(ctx context.Context, pool *pgxpool.Pool, groupID string) (models.GroupDetails, error) {
	log.Printf("[DB] Fetching group details: %s", groupID)

	var group models.GroupDetails

	// Fetch group basic information
	groupQuery := `SELECT group_id, group_name, description, created_by, extract(epoch from created_at)::bigint
		FROM groups
		WHERE group_id = $1`

	err := pool.QueryRow(ctx, groupQuery, groupID).Scan(
		&group.GroupID,
		&group.Name,
		&group.Description,
		&group.CreatedBy,
		&group.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		log.Printf("[DB] Group not found: %s", groupID)
		return models.GroupDetails{}, ErrGroupNotFound
	}
	if err != nil {
		return models.GroupDetails{}, NewDBError("GetGroup", err, "failed to query group")
	}

	// Fetch group members with user details
	membersQuery := `SELECT u.user_id, u.user_name, u.email, u.is_guest, extract(epoch from gm.joined_at)::bigint
		FROM group_members gm
		JOIN users u ON gm.user_id = u.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC`

	rows, err := pool.Query(ctx, membersQuery, groupID)
	if err != nil {
		return models.GroupDetails{}, NewDBError("GetGroup", err, "failed to query group members")
	}
	defer rows.Close()

	// Scan members into the group
	group.Members = make([]models.GroupUser, 0)
	for rows.Next() {
		var member models.GroupUser
		err := rows.Scan(&member.UserID, &member.Name, &member.Email, &member.Guest, &member.JoinedAt)
		if err != nil {
			return models.GroupDetails{}, NewDBError("GetGroup", err, "failed to scan member row")
		}
		group.Members = append(group.Members, member)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return models.GroupDetails{}, NewDBError("GetGroup", err, "error iterating member rows")
	}

	log.Printf("[DB] ✓ Group retrieved: %s with %d members", group.Name, len(group.Members))
	return group, nil
}

// AddGroupMembers adds multiple users to a group in a single batch operation.
// Uses batch operations for better performance when adding many members at once.
// Ignores duplicate memberships (ON CONFLICT DO NOTHING).
// Returns ErrInvalidInput if no user IDs are provided.
func AddGroupMembers(ctx context.Context, pool *pgxpool.Pool, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return ErrInvalidInput
	}

	log.Printf("[DB] Adding %d members to group: %s", len(userIDs), groupID)

	// Build batch queries for all users
	batch := &pgx.Batch{}
	insertQuery := `INSERT INTO group_members (user_id, group_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, group_id) DO NOTHING`

	now := time.Now()
	for _, userID := range userIDs {
		batch.Queue(insertQuery, userID, groupID, now)
	}

	// Execute batch
	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	// Check results for each query
	for i := range userIDs {
		_, err := br.Exec()
		if err != nil {
			return NewDBError("AddGroupMembers", err,
				fmt.Sprintf("failed to add member %d of %d", i+1, len(userIDs)))
		}
	}

	log.Printf("[DB] ✓ Successfully added %d members to group", len(userIDs))
	return nil
}

// AddGroupMember adds a single user to a group.
// This is a convenience function for adding one member at a time.
// Ignores duplicate memberships (ON CONFLICT DO NOTHING).
func AddGroupMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) error {
	log.Printf("[DB] Adding member %s to group %s", userID, groupID)

	query := `INSERT INTO group_members (user_id, group_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, group_id) DO NOTHING`

	_, err := pool.Exec(ctx, query, userID, groupID, time.Now())
	if err != nil {
		return NewDBError("AddGroupMember", err, "failed to add member")
	}

	log.Printf("[DB] ✓ Member added to group")
	return nil
}

// RemoveGroupMember removes a single user from a group.
// Note: The database will handle cascading deletes for related expenses if configured.
func RemoveGroupMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) error {
	log.Printf("[DB] Removing member %s from group %s", userID, groupID)

	query := `DELETE FROM group_members
		WHERE user_id = $1 AND group_id = $2`

	result, err := pool.Exec(ctx, query, userID, groupID)
	if err != nil {
		return NewDBError("RemoveGroupMember", err, "failed to remove member")
	}

	// Check if any rows were affected
	if result.RowsAffected() == 0 {
		log.Printf("[DB] Member not found in group")
		return ErrNotMember
	}

	log.Printf("[DB] ✓ Member removed from group")
	return nil
}

// RemoveGroupMembers removes multiple users from a group in a single batch operation.
// Uses batch operations for better performance when removing many members at once.
// Returns ErrInvalidInput if no user IDs are provided.
func RemoveGroupMembers(ctx context.Context, pool *pgxpool.Pool, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return ErrInvalidInput
	}

	log.Printf("[DB] Removing %d members from group: %s", len(userIDs), groupID)

	// Build batch queries for all users
	batch := &pgx.Batch{}
	deleteQuery := `DELETE FROM group_members
		WHERE user_id = $1 AND group_id = $2`

	for _, userID := range userIDs {
		batch.Queue(deleteQuery, userID, groupID)
	}

	// Execute batch
	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	// Check results for each query
	for i := range userIDs {
		_, err := br.Exec()
		if err != nil {
			return NewDBError("RemoveGroupMembers", err,
				fmt.Sprintf("failed to remove member %d of %d", i+1, len(userIDs)))
		}
	}

	log.Printf("[DB] ✓ Successfully removed %d members from group", len(userIDs))
	return nil
}
