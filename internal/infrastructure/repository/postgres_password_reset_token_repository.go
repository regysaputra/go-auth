package repository

import (
	"auth/internal/domain"
	"auth/internal/usecase"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresPasswordResetTokenRepository represents the Postgres password reset token repository object
type PostgresPasswordResetTokenRepository struct {
	db                    *pgxpool.Pool
	passwordResetTokenTTL time.Duration
}

// NewPostgresPasswordResetTokenRepository creates a new Postgres password reset token repository object
func NewPostgresPasswordResetTokenRepository(db *pgxpool.Pool, passwordResetTokenTTL time.Duration) usecase.PasswordResetTokenRepository {
	return &PostgresPasswordResetTokenRepository{
		db:                    db,
		passwordResetTokenTTL: passwordResetTokenTTL,
	}
}

// Save saves the password reset token to the database
func (r *PostgresPasswordResetTokenRepository) Save(ctx context.Context, userID int64, tokenHash string) error {
	expiresAt := time.Now().Add(r.passwordResetTokenTTL)
	sql := "INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(ctx, sql, userID, tokenHash, expiresAt)
	return err
}

// FindByToken finds the password reset token by token hash
func (r *PostgresPasswordResetTokenRepository) FindByToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	sql := "SELECT * FROM password_reset_tokens WHERE token_hash = $1 AND expires_at > NOW()"
	row := r.db.QueryRow(ctx, sql, tokenHash)

	var token domain.PasswordResetToken
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

// Delete deletes the password reset token by ID
func (r *PostgresPasswordResetTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sql := "DELETE FROM password_reset_tokens WHERE id = $1"
	_, err := r.db.Exec(ctx, sql, id)
	return err
}
