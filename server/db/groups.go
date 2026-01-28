// Package db provides database operations for group management.
// This file contains all group-related database operations including creating groups,
// managing group members, and retrieving group information.
package db

import (
	"context"
	"log"
	"time"

	"github.com/pranaovs/qashare/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateGroup creates a new group in the database and automatically adds the creator as a member.
// This operation is atomic - either both the group creation and membership addition succeed,
// or neither does (using a transaction).
// Takes a Group model with Name, Description, and CreatedBy populated, and adds GroupID and CreatedAt.
// Returns an error if the operation fails. The group's GroupID and CreatedAt fields will be populated upon success.
func CreateGroup(ctx context.Context, pool *pgxpool.Pool, group *models.Group) error {
	// Use WithTransaction helper for consistent transaction management
	err := WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Insert the group
		query := `INSERT INTO groups (group_name, description, created_by)
			VALUES ($1, $2, $3)
			RETURNING group_id, extract(epoch from created_at)::bigint`

		err := tx.QueryRow(ctx, query, group.Name, group.Description, group.CreatedBy).Scan(&group.GroupID, &group.CreatedAt)
		if err != nil {
			return err
		}

		// Add creator as the first member
		memberQuery := `INSERT INTO group_members (user_id, group_id, joined_at)
			VALUES ($1, $2, $3)`

		_, err = tx.Exec(ctx, memberQuery, group.CreatedBy, group.GroupID, time.Now())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// GetGroupCreator retrieves the user ID of the group creator.
// This is a lightweight query that only returns the creator ID, useful for authorization checks.
// Returns ErrNotFound if no group with the ID exists.
func GetGroupCreator(ctx context.Context, pool *pgxpool.Pool, groupID string) (string, error) {
	var creatorID string
	query := `SELECT created_by FROM groups WHERE group_id = $1`

	err := pool.QueryRow(ctx, query, groupID).Scan(&creatorID)
	if err == pgx.ErrNoRows {
		return "", ErrNotFound.Msgf("group with id %s not found", groupID)
	}
	if err != nil {
		return "", err
	}

	return creatorID, nil
}

// GetGroup retrieves complete group information including all members.
// Returns a models.GroupDetails struct with full details and a list of all group members.
// Returns ErrNotFound if no group with the ID exists.
func GetGroup(ctx context.Context, pool *pgxpool.Pool, groupID string) (models.GroupDetails, error) {
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
		return models.GroupDetails{}, ErrNotFound.Msgf("group with id %s not found", groupID)
	}
	if err != nil {
		return models.GroupDetails{}, err
	}

	// Fetch group members with user details
	membersQuery := `SELECT u.user_id, u.user_name, u.email, u.is_guest, extract(epoch from gm.joined_at)::bigint
		FROM group_members gm
		JOIN users u ON gm.user_id = u.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC`

	rows, err := pool.Query(ctx, membersQuery, groupID)
	if err != nil {
		return models.GroupDetails{}, err
	}
	defer rows.Close()

	// Scan members into the group
	group.Members = make([]models.GroupUser, 0)
	for rows.Next() {
		var member models.GroupUser
		err := rows.Scan(&member.UserID, &member.Name, &member.Email, &member.Guest, &member.JoinedAt)
		if err != nil {
			return models.GroupDetails{}, err
		}
		group.Members = append(group.Members, member)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return models.GroupDetails{}, err
	}

	return group, nil
}

// AddGroupMembers adds multiple users to a group in a single batch operation.
// Uses batch operations for better performance when adding many members at once.
// Ignores duplicate memberships (ON CONFLICT DO NOTHING).
// Returns ErrInvalidInput if no user IDs are provided.
func AddGroupMembers(ctx context.Context, pool *pgxpool.Pool, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return ErrInvalidInput.Msg("no user IDs provided")
	}

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
	defer func() {
		if err := br.Close(); err != nil {
			log.Printf("[DB] Error closing batch: %v", err)
		}
	}()

	// Check results for each query
	for range userIDs {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// AddGroupMember adds a single user to a group.
// This is a convenience function for adding one member at a time.
// Ignores duplicate memberships (ON CONFLICT DO NOTHING).
func AddGroupMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) error {
	query := `INSERT INTO group_members (user_id, group_id, joined_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, group_id) DO NOTHING`

	_, err := pool.Exec(ctx, query, userID, groupID, time.Now())
	if err != nil {
		return err
	}

	return nil
}

// RemoveGroupMember removes a single user from a group.
// Note: The database will handle cascading deletes for related expenses if configured.
func RemoveGroupMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) error {
	query := `DELETE FROM group_members
		WHERE user_id = $1 AND group_id = $2`

	result, err := pool.Exec(ctx, query, userID, groupID)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	if result.RowsAffected() == 0 {
		return ErrNotFound.Msgf("user %s is not a member of group %s", userID, groupID)
	}

	return nil
}

// RemoveGroupMembers removes multiple users from a group in a single batch operation.
// Uses batch operations for better performance when removing many members at once.
// Returns ErrInvalidInput if no user IDs are provided.
func RemoveGroupMembers(ctx context.Context, pool *pgxpool.Pool, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return ErrInvalidInput.Msg("no user IDs provided")
	}

	// Build batch queries for all users
	batch := &pgx.Batch{}
	deleteQuery := `DELETE FROM group_members
		WHERE user_id = $1 AND group_id = $2`

	for _, userID := range userIDs {
		batch.Queue(deleteQuery, userID, groupID)
	}

	// Execute batch
	br := pool.SendBatch(ctx, batch)

	defer func() {
		if err := br.Close(); err != nil {
			log.Printf("[DB] Error closing batch: %v", err)
		}
	}()

	// Check results for each query
	for range userIDs {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
