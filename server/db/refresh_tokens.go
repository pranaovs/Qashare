package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StoreToken inserts a refresh token record into the database.
func StoreToken(ctx context.Context, pool *pgxpool.Pool, tokenID, userID uuid.UUID, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (token_id, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := pool.Exec(ctx, query, tokenID, userID, expiresAt)
	return err
}

// DeleteRefreshToken removes a specific refresh token (used during rotation).
func DeleteRefreshToken(ctx context.Context, pool *pgxpool.Pool, tokenID uuid.UUID) error {
	result, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_id = $1`, tokenID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound.Msg("refresh token not found")
	}
	return nil
}

// DeleteUserRefreshTokens removes all refresh tokens for a user (used on logout/password change).
func DeleteUserRefreshTokens(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

// RefreshTokenExists checks if a refresh token exists and is not expired.
func RefreshTokenExists(ctx context.Context, pool *pgxpool.Pool, tokenID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM refresh_tokens WHERE token_id = $1 AND expires_at > NOW())`
	err := pool.QueryRow(ctx, query, tokenID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// DeleteExpiredTokens removes all expired refresh tokens from the database.
func DeleteExpiredTokens(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	result, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// StartTokenCleanup runs a background goroutine that periodically deletes expired refresh tokens.
// It stops when the context is cancelled.
func StartTokenCleanup(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("Token cleanup stopped")
				return
			case <-ticker.C:
				deleted, err := DeleteExpiredTokens(ctx, pool)
				if err != nil {
					slog.Error("Failed to clean up expired tokens", "error", err)
					continue
				}
				if deleted > 0 {
					slog.Info("Cleaned up expired refresh tokens", "count", deleted)
				}
			}
		}
	}()
}
