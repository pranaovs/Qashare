package db

import (
	"context"
	"errors"
	"time"

	"shared-expenses-app/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateUser inserts a new user into the database and returns the new user's ID.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, name, email, password string) (models.User, error) {
	// Check if user already exists
	_, err := GetUserFromEmail(ctx, pool, email)
	if err == nil {
		// User already exists
		return models.User{}, errors.New("user with this email already exists")
	} else if err != nil && err.Error() != "email not registered" {
		// Some other database error
		return models.User{}, err
	}

	// Add user to database
	var newUser models.User
	err = pool.QueryRow(
		ctx,
		`INSERT INTO users (user_name, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint`,
		name, email, password, time.Now(),
	).Scan(&newUser.UserID, &newUser.Name, &newUser.Email, &newUser.Guest, &newUser.CreatedAt)
	if err != nil {
		return models.User{}, err
	}
	// Return the new user's ID
	return newUser, nil
}

// GetUserFromEmail checks if a user with the given email exists.
// Returns (exists bool, userId string, err error). If user does not exist, userId will be empty string.
func GetUserFromEmail(ctx context.Context, pool *pgxpool.Pool, email string) (models.User, error) {
	var user models.User
	err := pool.QueryRow(ctx,
		`SELECT user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint
		FROM users
		WHERE email = $1`,
		email,
	).Scan(&user.UserID, &user.Name, &user.Email, &user.Guest, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.User{}, errors.New("email not registered") // email does not exist
	}
	if err != nil {
		return models.User{}, err // database error
	}

	return user, nil // user exists
}

func GetUserCredentials(ctx context.Context, pool *pgxpool.Pool, email string) (string, string, error) {
	var userID, passwordHash string
	err := pool.QueryRow(
		ctx,
		`SELECT user_id, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&userID, &passwordHash)
	if err == pgx.ErrNoRows {
		return "", "", errors.New("email not registered")
	}

	if err != nil {
		return "", "", err
	}

	return userID, passwordHash, nil
}

func GetUser(ctx context.Context, pool *pgxpool.Pool, userID string) (models.User, error) {
	var user models.User
	err := pool.QueryRow(
		ctx,
		`SELECT user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint FROM users WHERE user_id = $1`,
		userID,
	).Scan(&user.UserID, &user.Name, &user.Email, &user.Guest, &user.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.User{}, errors.New("user not found")
	}
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func UsersRelated(ctx context.Context, pool *pgxpool.Pool, userID1, userID2 string) (bool, error) {
	var areRelated bool
	err := pool.QueryRow(ctx, `
    SELECT EXISTS (
        SELECT 1
        FROM group_members gm1
        JOIN group_members gm2
        ON gm1.group_id = gm2.group_id
        WHERE gm1.user_id = $1
        AND gm2.user_id = $2
    )`, userID1, userID2).Scan(&areRelated)
	if err != nil {
		return false, err
	}

	return areRelated, nil
}

// AdminOfGroups return a list of models.Group where the user is the creator
func AdminOfGroups(ctx context.Context, pool *pgxpool.Pool, userID string) ([]models.Group, error) {
	rows, err := pool.Query(ctx, `
		SELECT group_id, group_name, description, created_by, extract(epoch from created_at)::bigint
		FROM groups
		WHERE created_by = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// MemberOfGroups returns the groups where the user is a member of (includes created groups)
func MemberOfGroups(ctx context.Context, pool *pgxpool.Pool, userID string) ([]models.Group, error) {
	rows, err := pool.Query(ctx, `
		SELECT g.group_id, g.group_name, g.description, g.created_by, extract(epoch from g.created_at)::bigint
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.group_id
		WHERE gm.user_id = $1
		ORDER BY g.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// UserExists checks if a user with the given userID exists in the database.
func UserExists(ctx context.Context, pool *pgxpool.Pool, userID string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT true FROM users WHERE user_id = $1`,
		userID,
	).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func MemberOfGroup(ctx context.Context, pool *pgxpool.Pool, userID, groupID string) (bool, error) {
	var isMember bool
	err := pool.QueryRow(ctx,
		`SELECT true FROM group_members WHERE user_id = $1 AND group_id = $2`,
		userID, groupID,
	).Scan(&isMember)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
