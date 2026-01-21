// Package db provides database operations for user management.
// This file contains all user-related database operations including CRUD operations,
// user verification, and relationship checking between users.
package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pranaovs/qashare/models"
	"github.com/pranaovs/qashare/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateUser inserts a new user into the database and returns the newly created user's ID.
// Parameters:
//   - name: The user's display name
//   - email: The user's email address (must be unique)
//   - password: The hashed password for the user
//
// Returns the newly created user's ID or an error if the operation fails.
// Returns ErrEmailAlreadyExists if a user with the email already exists.
func CreateUser(ctx context.Context, pool *pgxpool.Pool, name, email, password string) (string, error) {
	log.Printf("[DB] Creating new user with email: %s", email)

	// Check if user already exists with this email
	_, err := GetUserFromEmail(ctx, pool, email)
	if err == nil {
		// User already exists
		log.Printf("[DB] User creation failed: email already exists: %s", email)
		return "", ErrEmailAlreadyExists
	} else if err != nil && err != ErrEmailNotRegistered {
		// Some other database error occurred
		return "", NewDBError("CreateUser", err, "failed to check existing user")
	}

	// Insert the new user into the database
	var userID string
	query := `INSERT INTO users (user_name, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id`
	
	err = pool.QueryRow(ctx, query, name, email, password, time.Now()).Scan(&userID)
	if err != nil {
		// Check for duplicate key violation (race condition)
		if IsDuplicateKey(err) {
			return "", ErrEmailAlreadyExists
		}
		return "", NewDBError("CreateUser", err, "failed to insert user")
	}

	log.Printf("[DB] ✓ User created successfully with ID: %s", userID)
	return userID, nil
}

// GetUserFromEmail retrieves a user by their email address.
// This is commonly used for login and authentication purposes.
// Returns ErrEmailNotRegistered if no user with the email exists.
func GetUserFromEmail(ctx context.Context, pool *pgxpool.Pool, email string) (models.User, error) {
	log.Printf("[DB] Fetching user by email: %s", email)

	var user models.User
	query := `SELECT user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint
		FROM users
		WHERE email = $1`
	
	err := pool.QueryRow(ctx, query, email).Scan(
		&user.UserID, &user.Name, &user.Email, &user.Guest, &user.CreatedAt,
	)
	
	if err == pgx.ErrNoRows {
		log.Printf("[DB] User not found for email: %s", email)
		return models.User{}, ErrEmailNotRegistered
	}
	if err != nil {
		return models.User{}, NewDBError("GetUserFromEmail", err, "failed to query user")
	}

	log.Printf("[DB] ✓ User found with ID: %s", user.UserID)
	return user, nil
}

// GetUserCredentials retrieves the user ID and password hash for authentication.
// This function is specifically designed for login verification.
// Only returns the minimal information needed for authentication.
// Returns ErrEmailNotRegistered if no user with the email exists.
func GetUserCredentials(ctx context.Context, pool *pgxpool.Pool, email string) (string, string, error) {
	log.Printf("[DB] Fetching credentials for email: %s", email)

	var userID, passwordHash string
	query := `SELECT user_id, password_hash FROM users WHERE email = $1`
	
	err := pool.QueryRow(ctx, query, email).Scan(&userID, &passwordHash)
	if err == pgx.ErrNoRows {
		log.Printf("[DB] Credentials not found for email: %s", email)
		return "", "", ErrEmailNotRegistered
	}
	if err != nil {
		return "", "", NewDBError("GetUserCredentials", err, "failed to query credentials")
	}

	log.Printf("[DB] ✓ Credentials retrieved for user ID: %s", userID)
	return userID, passwordHash, nil
}

// GetUser retrieves a user by their unique user ID.
// Returns ErrUserNotFound if no user with the ID exists.
func GetUser(ctx context.Context, pool *pgxpool.Pool, userID string) (models.User, error) {
	log.Printf("[DB] Fetching user by ID: %s", userID)

	var user models.User
	query := `SELECT user_id, user_name, email, is_guest, extract(epoch from created_at)::bigint 
		FROM users 
		WHERE user_id = $1`
	
	err := pool.QueryRow(ctx, query, userID).Scan(
		&user.UserID, &user.Name, &user.Email, &user.Guest, &user.CreatedAt,
	)
	
	if err == pgx.ErrNoRows {
		log.Printf("[DB] User not found with ID: %s", userID)
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, NewDBError("GetUser", err, "failed to query user")
	}

	log.Printf("[DB] ✓ User retrieved: %s", user.Name)
	return user, nil
}

// UsersRelated checks if two users are related through group membership.
// Two users are considered related if they share at least one group.
// This is useful for privacy checks to ensure users can only see information
// about other users they're connected to through groups.
// Returns nil if users are related, or ErrUsersNotRelated if not.
func UsersRelated(ctx context.Context, pool *pgxpool.Pool, userID1, userID2 string) error {
	log.Printf("[DB] Checking if users are related: %s and %s", userID1, userID2)

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
		return NewDBError("UsersRelated", err, "failed to check user relationship")
	}

	if !areRelated {
		log.Printf("[DB] Users are not related: %s and %s", userID1, userID2)
		return ErrUsersNotRelated
	}

	log.Printf("[DB] ✓ Users are related through shared groups")
	return nil
}

// AdminOfGroups returns all groups where the user is the creator/administrator.
// Groups are returned in descending order by creation date (newest first).
// This is useful for showing users the groups they manage.
func AdminOfGroups(ctx context.Context, pool *pgxpool.Pool, userID string) ([]models.Group, error) {
	log.Printf("[DB] Fetching groups where user is admin: %s", userID)

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

	log.Printf("[DB] ✓ Found %d admin groups for user", len(groups))
	return groups, nil
}

// MemberOfGroups returns all groups where the user is a member.
// This includes both groups the user created and groups they were added to.
// Groups are returned in descending order by creation date (newest first).
func MemberOfGroups(ctx context.Context, pool *pgxpool.Pool, userID string) ([]models.Group, error) {
	log.Printf("[DB] Fetching groups where user is member: %s", userID)

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

	log.Printf("[DB] ✓ Found %d member groups for user", len(groups))
	return groups, nil
}

// UserExists checks if a user with the given ID exists in the database.
// This is a lightweight check that doesn't retrieve the full user record.
// Returns nil if user exists, or ErrUserNotFound if not.
func UserExists(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	log.Printf("[DB] Checking if user exists: %s", userID)

	exists, err := RecordExists(ctx, pool, "users", "user_id = $1", userID)
	if err != nil {
		return NewDBError("UserExists", err, "failed to check user existence")
	}

	if !exists {
		log.Printf("[DB] User does not exist: %s", userID)
		return ErrUserNotFound
	}

	log.Printf("[DB] ✓ User exists: %s", userID)
	return nil
}

// MemberOfGroup checks if a user is a member of a specific group.
// This is used for authorization checks before allowing group operations.
// Returns nil if user is a member, or ErrNotMember if not.
func MemberOfGroup(ctx context.Context, pool *pgxpool.Pool, userID, groupID string) error {
	log.Printf("[DB] Checking if user %s is member of group %s", userID, groupID)

	exists, err := RecordExists(ctx, pool, "group_members", 
		"user_id = $1 AND group_id = $2", userID, groupID)
	if err != nil {
		return NewDBError("MemberOfGroup", err, "failed to check group membership")
	}

	if !exists {
		log.Printf("[DB] User is not a member of group: %s", groupID)
		return ErrNotMember
	}

	log.Printf("[DB] ✓ User is a member of group")
	return nil
}

// AllMembersOfGroup verifies that all users in the provided list are members of the group.
// This is useful for validating expense splits where all participants must be group members.
// Returns nil if all users are members, or ErrNotMember if any user is not a member.
func AllMembersOfGroup(ctx context.Context, pool *pgxpool.Pool, userIDs []string, groupID string) error {
	if len(userIDs) == 0 {
		return nil
	}

	log.Printf("[DB] Checking if %d users are members of group %s", len(userIDs), groupID)

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
		log.Printf("[DB] Not all users are members: expected %d, got %d", len(uniqueUserIDs), count)
		return ErrNotMember
	}

	log.Printf("[DB] ✓ All %d users are members of the group", count)
	return nil
}
