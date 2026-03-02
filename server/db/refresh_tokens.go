package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StoreToken inserts a refresh token record into the database.
func StoreToken(ctx context.Context, pool *pgxpool.Pool, tokenID, userID uuid.UUID, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (token_id, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := pool.Exec(ctx, query, tokenID, userID, expiresAt)
	return err
}

// DeleteToken removes a specific refresh token (e.g., for logout or revocation).
func DeleteToken(ctx context.Context, pool *pgxpool.Pool, tokenID uuid.UUID) error {
	result, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_id = $1`, tokenID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound.Msg("refresh token not found")
	}
	return nil
}

// RotateToken atomically deletes the old refresh token and inserts a new one.
// Returns ErrNotFound if the old token doesn't exist (already used or revoked).
func RotateToken(ctx context.Context, pool *pgxpool.Pool, oldTokenID, newTokenID, userID uuid.UUID, newExpiresAt time.Time) error {
	return WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_id = $1`, oldTokenID)
		if err != nil {
			return err
		}
		if result.RowsAffected() == 0 {
			return ErrNotFound.Msg("refresh token not found")
		}

		_, err = tx.Exec(ctx, `INSERT INTO refresh_tokens (token_id, user_id, expires_at) VALUES ($1, $2, $3)`, newTokenID, userID, newExpiresAt)
		return err
	})
}

// DeleteTokens removes all refresh tokens for a user (used on logout/password change).
func DeleteTokens(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

// TokenExists checks if a refresh token exists and is not expired.
func TokenExists(ctx context.Context, pool *pgxpool.Pool, tokenID uuid.UUID) (bool, error) {
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
// It stops when the context is cancelled. The returned channel is closed once the goroutine exits.
func StartTokenCleanup(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
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
	return done
}
