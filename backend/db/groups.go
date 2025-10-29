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
func CreateGroup(ctx context.Context, pool *pgxpool.Pool, name, description, ownerUserID string) (models.Group, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.Group{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var group models.Group

	err = tx.QueryRow(
		ctx,
		`INSERT INTO groups (group_name, description, created_by, created_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING group_id, group_name, description, created_by, extract(epoch from created_at)::bigint`,
		name, description, ownerUserID, time.Now(),
	).Scan(&group.GroupID, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt)
	if err != nil {
		return models.Group{}, err
	}

	_, err = tx.Exec(
		ctx,
		`INSERT INTO group_members (user_id, group_id, joined_at)
		 VALUES ($1, $2, $3)`,
		ownerUserID, group.GroupID, time.Now(),
	)
	if err != nil {
		return models.Group{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Group{}, err
	}

	return group, nil
}

func GetGroup(ctx context.Context, pool *pgxpool.Pool, groupID string) (models.Group, error) {
	var group models.Group

	err := pool.QueryRow(
		ctx,
		`SELECT name, description, created_by, extract(epoch from created_at)::bigint
		FROM groups
		WHERE group_id = $1`,
		groupID,
	).Scan(&group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.Group{}, errors.New("group not found")
	}
	if err != nil {
		return models.Group{}, err
	}

	// Fetch group members
	rows, err := pool.Query(
		ctx,
		`SELECT u.user_id, u.user_name, u.email, u.is_guest, gm.joined_at
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

	validUserIDs := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		exists, err := UserExists(ctx, pool, userID)
		if err != nil {
			return err
		}
		if exists {
			validUserIDs = append(validUserIDs, userID)
		}
	}

	batch := &pgx.Batch{}
	for _, userID := range validUserIDs {
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
