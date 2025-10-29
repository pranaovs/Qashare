package db

import (
	"context"
	"errors"
	"time"

	"shared-expenses-app/models"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateUser inserts a new user into the database and returns the new user's ID.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, name, email, password string) (string, error) {
	// Check if user already exists
	ok, _, err := GetUserIDFromEmail(context.Background(), pool, email)
	if err != nil {
		return "", err
	}

	if ok {
		return "", errors.New("user already exists")
	}

	// Add user to database
	var userID string
	err = pool.QueryRow(
		context.Background(),
		`INSERT INTO users (user_name, email, password_hash, created_at)
			 VALUES ($1, $2, $3, $4)
			 RETURNING user_id`,
		name, email, password, time.Now(),
	).Scan(&userID)
	if err != nil {
		return "", err
	}
	// Return the new user's ID
	return userID, nil
}

// GetUserIDFromEmail checks if a user with the given email exists.
// Returns (exists bool, userId string, err error). If user does not exist, userId will be empty string.
func GetUserIDFromEmail(ctx context.Context, pool *pgxpool.Pool, email string) (bool, string, error) {
	var userID string
	err := pool.QueryRow(ctx,
		`SELECT user_id FROM users WHERE email = $1`,
		email,
	).Scan(&userID)
	if err == pgx.ErrNoRows {
		return false, "", nil // user not found
	}
	if err != nil {
		return false, "", err // database error
	}

	return true, userID, nil // user exists
}

func GetUserCredentials(ctx context.Context, pool *pgxpool.Pool, email string) (string, string, error) {
	var userID, passwordHash string
	err := pool.QueryRow(
		context.Background(),
		`select user_id, password_hash from users where email = $1`,
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
		context.Background(),
		`select user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint from users where user_id = $1`,
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
