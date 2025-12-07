package repository

import (
	"auth/internal/domain"
	"auth/internal/usecase"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresVerificationTokenRepository represents the Postgres verification token repository object
type PostgresVerificationTokenRepository struct {
	db                   *pgxpool.Pool
	verificationTokenTTL time.Duration
}

// NewPostgresVerificationTokenRepository creates a new Postgres verification token repository object
func NewPostgresVerificationTokenRepository(db *pgxpool.Pool, verificationTokenTTL time.Duration) usecase.VerificationTokenRepository {
	return &PostgresVerificationTokenRepository{
		db:                   db,
		verificationTokenTTL: verificationTokenTTL,
	}
}

// Save saves the verification token to the database
func (r *PostgresVerificationTokenRepository) Save(ctx context.Context, userID int64, tokenHash string) error {
	expiresAt := time.Now().Add(r.verificationTokenTTL)
	sql := `INSERT INTO verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, sql, userID, tokenHash, expiresAt)
	return err
}

// FindByToken finds the verification token by token hash
func (r *PostgresVerificationTokenRepository) FindByToken(ctx context.Context, tokenHash string) (*domain.VerificationToken, error) {
	sql := `SELECT * FROM verification_tokens WHERE token_hash = $1 AND expires_at > NOW()`
	var token domain.VerificationToken
	row := r.db.QueryRow(ctx, sql, tokenHash)
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Delete deletes the verification token by ID
func (r *PostgresVerificationTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `DELETE FROM verification_tokens WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}
