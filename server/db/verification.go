package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VerifyEmail looks up the verification token, checks expiry, sets email_verified=true,
// and deletes all verification tokens for the user.
// Returns ErrNotFound if the token doesn't exist, or ErrExpiredToken if it has expired.
func VerifyEmail(ctx context.Context, pool *pgxpool.Pool, token uuid.UUID) error {
	return WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
		var userID uuid.UUID
		var expiresAt time.Time

		err := tx.QueryRow(ctx,
			`SELECT user_id, expires_at FROM email_verification_tokens WHERE token = $1 FOR UPDATE`,
			token,
		).Scan(&userID, &expiresAt)

		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		if err != nil {
			return err
		}

		if time.Now().After(expiresAt) {
			_, _ = tx.Exec(ctx, `DELETE FROM email_verification_tokens WHERE token = $1`, token)
			return ErrExpiredToken
		}

		_, err = tx.Exec(ctx, `UPDATE users SET email_verified = true WHERE user_id = $1`, userID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID)
		return err
	})
}

// DeleteExpiredVerificationTokens removes all expired verification tokens.
func DeleteExpiredVerificationTokens(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	result, err := pool.Exec(ctx, `DELETE FROM email_verification_tokens WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
