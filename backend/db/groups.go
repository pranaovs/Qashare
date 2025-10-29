package db

import (
	"context"
	"errors"
	"time"

	"shared-expenses-app/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateGroup(ctx context.Context, pool *pgxpool.Pool, name, description, ownerUserID string) (models.Group, error) {
	var group models.Group

	err := pool.QueryRow(
		ctx,
		`INSERT INTO groups (group_name, description, created_by, created_at)
			 VALUES ($1, $2, $3, $4)
			 RETURNING group_id, group_name, description, created_by, extract(epoch from created_at)::bigint`,
		name, description, ownerUserID, time.Now(),
	).Scan(&group.GroupID, &group.Name, &group.Description, &group.CreatedBy, &group.CreatedAt)
	if err != nil {
		return models.Group{}, err
	}

	// Return the new user's ID
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
