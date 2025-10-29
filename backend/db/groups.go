package db

import (
	"context"
	"errors"
	"time"

	"shared-expenses-app/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateGroup inserts a new group into the database and adds the owner as a member.
func CreateGroup(ctx context.Context, pool *pgxpool.Pool, name, description, ownerUserID string) (string, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var groupID string

	err = tx.QueryRow(
		ctx,
		`INSERT INTO groups (group_name, description, created_by, created_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING group_id`,
		name, description, ownerUserID, time.Now(),
	).Scan(&groupID)
	if err != nil {
		return "", err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO group_members (user_id, group_id, joined_at)
		 VALUES ($1, $2, $3)`,
		ownerUserID, groupID, time.Now(),
	)
	if err != nil {
		return "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return "", err
	}

	return groupID, nil
}

// GetGroupCreator returns only the creator ID of a group.
func GetGroupCreator(ctx context.Context, pool *pgxpool.Pool, groupID string) (string, error) {
	var creatorID string
	err := pool.QueryRow(
		ctx,
		`SELECT created_by FROM groups WHERE group_id = $1`,
		groupID,
	).Scan(&creatorID)
	if err == pgx.ErrNoRows {
		return "", errors.New("group not found")
	}
	if err != nil {
		return "", err
	}
	return creatorID, nil
}

func GetGroup(ctx context.Context, pool *pgxpool.Pool, groupID string) (models.Group, error) {
	var group models.Group

	err := pool.QueryRow(
		ctx,
		`SELECT group_id, group_name, description, created_by, extract(epoch from created_at)::bigint
		FROM groups
		WHERE group_id = $1`,
		groupID,
	).Scan(&group.GroupID, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.Group{}, errors.New("group not found")
	}
	if err != nil {
		return models.Group{}, err
	}

	// Fetch group members
	rows, err := pool.Query(
		ctx,
		`SELECT u.user_id, u.user_name, u.email, u.is_guest, extract(epoch from gm.joined_at)::bigint
		 FROM group_members gm
		 JOIN users u ON gm.user_id = u.user_id
		 WHERE gm.group_id = $1`,
		groupID,
	)
	if err != nil {
		return models.Group{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var member models.GroupUser
		err := rows.Scan(&member.UserID, &member.Name, &member.Email, &member.Guest, &member.JoinedAt)
		if err != nil {
			return models.Group{}, err
		}
		group.Members = append(group.Members, member)
	}

	return group, nil
}

// AddGroupMembers adds multiple users to a group.
func AddGroupMembers(ctx context.Context, pool *pgxpool.Pool, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return errors.New("no user IDs provided")
	}

	batch := &pgx.Batch{}
	for _, userID := range userIDs {
		batch.Queue(
			`INSERT INTO group_members (user_id, group_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			userID, groupID,
		)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for range userIDs {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// AddGroupMember adds a single user to a group
func AddGroupMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) error {
	_, err := pool.Exec(
		ctx,
		`INSERT INTO group_members (user_id, group_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`,
		userID, groupID)
	if err != nil {
		return err
	}
	return nil
}

func RemoveGroupMember(ctx context.Context, pool *pgxpool.Pool, groupID, userID string) error {
	_, err := pool.Exec(
		ctx,
		`DELETE FROM group_members
	WHERE user_id = $1 AND group_id = $2`,
		userID, groupID)
	if err == pgx.ErrNoRows {
		return errors.New("member not found in group")
	}
	if err != nil {
		return err
	}
	return nil
}

// RemoveGroupMembers removes multiple users from a group.
func RemoveGroupMembers(ctx context.Context, pool *pgxpool.Pool, groupID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return errors.New("no user IDs provided")
	}

	batch := &pgx.Batch{}
	for _, userID := range userIDs {
		batch.Queue(
			`DELETE FROM group_members
			 WHERE user_id = $1 AND group_id = $2`,
			userID, groupID,
		)
	}

	br := pool.SendBatch(ctx, batch)
	defer br.Close()

	for range userIDs {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
