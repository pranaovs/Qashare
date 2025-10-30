package db

import (
	"context"
	"errors"
	"time"

	"shared-expenses-app/models"
	"shared-expenses-app/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Sentinel errors for user-related operations
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrNotMember       = errors.New("not a member")
	ErrUsersNotRelated = errors.New("users not related")
)

// CreateUser inserts a new user into the database and returns the newly created user's ID.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, name, email, password string) (string, error) {
	// Check if user already exists
	_, err := GetUserFromEmail(ctx, pool, email)
	if err == nil {
		// User already exists
		return "", errors.New("user with this email already exists")
	} else if err != nil && err.Error() != "email not registered" {
		// Some other database error
		return "", err
	}

	// Add user to database
	var userID string
	err = pool.QueryRow(
		ctx,
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

// GetUserFromEmail retrieves the user with the given email.
// Returns (models.User, error). If the user does not exist, returns an empty models.User and an error.
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

// UsersRelated checks if two users are related (share at least one group).
// Returns nil if users are related, or ErrUsersNotRelated if not.
func UsersRelated(ctx context.Context, pool *pgxpool.Pool, userID1, userID2 string) error {
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
		return err
	}

	if !areRelated {
		return ErrUsersNotRelated
	}

	return nil
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

	groups := make([]models.Group, 0)
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

	groups := make([]models.Group, 0)
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
// Returns nil if user exists, or ErrUserNotFound if not.
func UserExists(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT true FROM users WHERE user_id = $1`,
		userID,
	).Scan(&exists)
	if err == pgx.ErrNoRows {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	return nil
}

// MemberOfGroup checks if a user is a member of a group.
// Returns nil if user is a member, or ErrNotMember if not.
func MemberOfGroup(ctx context.Context, pool *pgxpool.Pool, userID, groupID string) error {
	var isMember bool
	err := pool.QueryRow(ctx,
		`SELECT true FROM group_members WHERE user_id = $1 AND group_id = $2`,
		userID, groupID,
	).Scan(&isMember)
	if err == pgx.ErrNoRows {
		return ErrNotMember
	}
	if err != nil {
		return err
	}

	return nil
}

// AllMembersOfGroup checks if all users in the provided userIDs slice are members of the group.
// Returns nil if all users are members, or an error if any user is not a member.
func AllMembersOfGroup(ctx context.Context, pool *pgxpool.Pool, userIDs []string, groupID string) error {
	if len(userIDs) == 0 {
		return nil
	}

	uniqueUserIDs := utils.GetUniqueUserIDs(userIDs)

	// Query to count how many of the provided userIDs are actually members
	var count int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT user_id)
		 FROM group_members
		 WHERE group_id = $1 AND user_id = ANY($2)`,
		groupID, uniqueUserIDs,
	).Scan(&count)
	if err != nil {
		return err
	}

	// If count doesn't match the number of userIDs, some users are not members
	if count != len(uniqueUserIDs) {
		return ErrNotMember
	}

	return nil
}
