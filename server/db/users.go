// Package db provides database operations for user management.
// This file contains all user-related database operations including CRUD operations,
// user verification, and relationship checking between users.
package db

import (
	"context"
	"strings"

	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateUser inserts a new non-guest (fully authenticated) user into the database.
// Guest accounts should normally be created using CreateGuest. If an existing guest user
// is found for the given email (signaled by ErrUserIsGuest from GetUserFromEmail), this
// function will handle that case when creating the full user account.
// Takes a User model with Name, Email, and PasswordHash populated, and adds UserID and CreatedAt.
// Returns ErrEmailAlreadyExists if a non-guest user with the email already exists.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, user *models.User) error {
	// Check if user already exists with this email
	_, err := GetUserFromEmail(ctx, pool, user.Email)

	var query string
	user.Guest = false

	query = `INSERT INTO users (user_name, email, password_hash, is_guest)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, extract(epoch from created_at)::bigint`

	if err == nil {
		// User already exists
		return ErrEmailAlreadyExists
	} else if err == ErrUserIsGuest {
		// Update the existing guest user to become a regular user
		query = `UPDATE users
			SET user_name = $1, password_hash = $3, is_guest = $4, created_at = NOW()
			WHERE email = $2
			RETURNING user_id, extract(epoch from created_at)::bigint`
	} else if err != ErrEmailNotRegistered {
		// Some other database error occurred
		return NewDBError("CreateUser", err, "failed to check existing user")
	}
	err = pool.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash, user.Guest).Scan(&user.UserID, &user.CreatedAt)
	if err != nil {
		// Check for duplicate key violation (race condition)
		if IsDuplicateKey(err) {
			return ErrEmailAlreadyExists
		}
		return NewDBError("CreateUser", err, "failed to insert user")
	}

	user.PasswordHash = nil // Remove password hash after insertion
	return nil
}

// CreateGuest inserts a new guest user into the database.
// The guest user is identified by email and has no password. The user name is derived
// from the part of the email before the "@" symbol. This function also records which
// existing user added the guest in the guests table.
// Takes a context, a database connection pool, the guest's email address, and the
// user ID of the user who added the guest.
// Returns the created User model with UserID and CreatedAt populated.
// Returns ErrEmailAlreadyExists if a user with the given email already exists.
// May return a DBError (via NewDBError) if any database operation fails.
func CreateGuest(ctx context.Context, pool *pgxpool.Pool, email string, addedBy string) (models.User, error) {
	// Check if user already exists with this email
	existing, err := GetUserFromEmail(ctx, pool, email)
	if err == nil {
		// User already exists
		return models.User{}, ErrEmailAlreadyExists
	} else if err == ErrUserIsGuest {
		return existing, nil // Guest user already exists, return it
	} else if err != ErrEmailNotRegistered {
		// Some other database error occurred
		return models.User{}, NewDBError("CreateGuest", err, "failed to check existing user")
	}

	var user models.User
	user.Email = email
	// Set guest user name as the part before the "@" in the email
	user.Name, _, _ = strings.Cut(email, "@")
	user.Guest = true

	err = WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		// Insert the guest user
		query := `INSERT INTO users (user_name, email, is_guest)
			VALUES ($1, $2, $3)
			RETURNING user_id, extract(epoch from created_at)::bigint`

		err := tx.QueryRow(ctx, query, user.Name, user.Email, user.Guest).Scan(&user.UserID, &user.CreatedAt)
		if err != nil {
			// Check for duplicate key violation (race condition)
			if IsDuplicateKey(err) {
				return ErrEmailAlreadyExists
			}
			return NewDBError("CreateGuest", err, "failed to insert guest user")
		}

		// Record who added this guest user
		query = `INSERT INTO guests (user_id, added_by)
			VALUES ($1, $2)`

		_, err = tx.Exec(ctx, query, user.UserID, addedBy)
		if err != nil {
			return NewDBError("CreateGuest", err, "failed to record guest addition")
		}

		return nil
	})
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

// GetUserFromEmail retrieves a user by their email address.
// This is commonly used for login and authentication purposes.
// Returns ErrEmailNotRegistered if no user with the email exists.
func GetUserFromEmail(ctx context.Context, pool *pgxpool.Pool, email string) (models.User, error) {
	var user models.User
	query := `SELECT user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint
		FROM users
		WHERE email = $1`

	err := pool.QueryRow(ctx, query, email).Scan(
		&user.UserID, &user.Name, &user.Email, &user.Guest, &user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return models.User{}, ErrEmailNotRegistered
	}
	if err != nil {
		return models.User{}, NewDBError("GetUserFromEmail", err, "failed to query user")
	}

	return user, nil
}

// GetUserCredentials retrieves the user ID and password hash for authentication.
// This function is specifically designed for login verification.
// Only returns the minimal information needed for authentication.
// Returns ErrEmailNotRegistered if no user with the email exists.
func GetUserCredentials(ctx context.Context, pool *pgxpool.Pool, email string) (string, string, error) {
	var userID, passwordHash string
	query := `SELECT user_id, password_hash FROM users WHERE email = $1`

	err := pool.QueryRow(ctx, query, email).Scan(&userID, &passwordHash)
	if err == pgx.ErrNoRows {
		return "", "", ErrEmailNotRegistered
	}
	if err != nil {
		return "", "", NewDBError("GetUserCredentials", err, "failed to query credentials")
	}

	return userID, passwordHash, nil
}

// GetUser retrieves a user by their unique user ID.
// Returns ErrUserNotFound if no user with the ID exists.
func GetUser(ctx context.Context, pool *pgxpool.Pool, userID string) (models.User, error) {
	var user models.User
	query := `SELECT user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint 
		FROM users
		WHERE user_id = $1`

	err := pool.QueryRow(ctx, query, userID).Scan(
		&user.UserID, &user.Name, &user.Email, &user.Guest, &user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, NewDBError("GetUser", err, "failed to query user")
	}

	return user, nil
}

// UsersRelated checks if two users are related through group membership.
// Two users are considered related if they share at least one group.
// This is useful for privacy checks to ensure users can only see information
// about other users they're connected to through groups.
// Returns true if users are related, false otherwise, and an error if the check fails.
func UsersRelated(ctx context.Context, pool *pgxpool.Pool, userID1, userID2 string) (bool, error) {
	// Query to check if users share at least one group
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM group_members gm1
			JOIN group_members gm2 ON gm1.group_id = gm2.group_id
			WHERE gm1.user_id = $1 AND gm2.user_id = $2
		)`

	var areRelated bool
	err := pool.QueryRow(ctx, query, userID1, userID2).Scan(&areRelated)
	if err != nil {
		return false, NewDBError("UsersRelated", err, "failed to check user relationship")
	}

	return areRelated, nil
}

// AdminOfGroups returns all groups where the user is the creator/administrator.
// Groups are returned in descending order by creation date (newest first).
// This is useful for showing users the groups they manage.
func AdminOfGroups(ctx context.Context, pool *pgxpool.Pool, userID string) ([]models.Group, error) {
	query := `
		SELECT group_id, group_name, description, created_by, extract(epoch from created_at)::bigint
		FROM groups
		WHERE created_by = $1
		ORDER BY created_at DESC`

	rows, err := pool.Query(ctx, query, userID)
	if err != nil {
		return nil, NewDBError("AdminOfGroups", err, "failed to query admin groups")
	}
	defer rows.Close()

	// Scan results into groups slice
	var groups []models.Group
	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt)
		if err != nil {
			return nil, NewDBError("AdminOfGroups", err, "failed to scan group row")
		}
		groups = append(groups, g)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, NewDBError("AdminOfGroups", err, "error iterating group rows")
	}

	return groups, nil
}

// MemberOfGroups returns all groups where the user is a member.
// This includes both groups the user created and groups they were added to.
// Groups are returned in descending order by creation date (newest first).
func MemberOfGroups(ctx context.Context, pool *pgxpool.Pool, userID string) ([]models.Group, error) {
	query := `
		SELECT g.group_id, g.group_name, g.description, g.created_by, extract(epoch from g.created_at)::bigint
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.group_id
		WHERE gm.user_id = $1
		ORDER BY g.created_at DESC`

	rows, err := pool.Query(ctx, query, userID)
	if err != nil {
		return nil, NewDBError("MemberOfGroups", err, "failed to query member groups")
	}
	defer rows.Close()

	// Scan results into groups slice
	var groups []models.Group
	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.GroupID, &g.Name, &g.Description, &g.CreatedBy, &g.CreatedAt)
		if err != nil {
			return nil, NewDBError("MemberOfGroups", err, "failed to scan group row")
		}
		groups = append(groups, g)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, NewDBError("MemberOfGroups", err, "error iterating group rows")
	}

	return groups, nil
}

// UserExists checks if a user with the given ID exists in the database.
// This is a lightweight check that doesn't retrieve the full user record.
// Returns nil if user exists, or ErrUserNotFound if not.
func UserExists(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	exists, err := RecordExists(ctx, pool, "users", "user_id = $1", userID)
	if err != nil {
		return NewDBError("UserExists", err, "failed to check user existence")
	}

	if !exists {
		return ErrUserNotFound
	}

	return nil
}

// MemberOfGroup checks if a user is a member of a specific group.
// This is used for authorization checks before allowing group operations.
// Returns nil if user is a member, or ErrNotMember if not.
func MemberOfGroup(ctx context.Context, pool *pgxpool.Pool, userID, groupID string) error {
	exists, err := RecordExists(ctx, pool, "group_members",
		"user_id = $1 AND group_id = $2", userID, groupID)
	if err != nil {
		return NewDBError("MemberOfGroup", err, "failed to check group membership")
	}

	if !exists {
		return ErrNotMember
	}

	return nil
}

// AllMembersOfGroup verifies that all users in the provided list are members of the group.
// This is useful for validating expense splits where all participants must be group members.
// Returns nil if all users are members, or ErrNotMember if any user is not a member.
func AllMembersOfGroup(ctx context.Context, pool *pgxpool.Pool, userIDs []string, groupID string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Get unique user IDs to avoid checking duplicates
	uniqueUserIDs := utils.GetUniqueUserIDs(userIDs)

	// Count how many of the provided user IDs are actually members
	query := `SELECT COUNT(DISTINCT user_id)
		FROM group_members
		WHERE group_id = $1 AND user_id = ANY($2)`

	var count int
	err := pool.QueryRow(ctx, query, groupID, uniqueUserIDs).Scan(&count)
	if err != nil {
		return NewDBError("AllMembersOfGroup", err, "failed to count group members")
	}

	// If count doesn't match, some users are not members
	if count != len(uniqueUserIDs) {
		return ErrNotMember
	}

	return nil
}
