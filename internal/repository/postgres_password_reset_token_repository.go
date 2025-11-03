package repository

import (
	"auth/internal/domain"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresPasswordResetTokenRepository represents the Postgres password reset token repository object
type PostgresPasswordResetTokenRepository struct {
	db *pgxpool.Pool
}

// NewPostgresPasswordResetTokenRepository creates a new Postgres password reset token repository object
func NewPostgresPasswordResetTokenRepository(db *pgxpool.Pool) *PostgresPasswordResetTokenRepository {
	return &PostgresPasswordResetTokenRepository{
		db: db,
	}
}

// Generate generates a random string of length 16
func (r *PostgresPasswordResetTokenRepository) Generate() (string, error) {
	// Generate 32 bytes of random data
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to a URL-safe base64 string
	return base64.URLEncoding.EncodeToString(b), nil
}

// Hash hashes the given token
func (r *PostgresPasswordResetTokenRepository) Hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// Save saves the password reset token to the database
func (r *PostgresPasswordResetTokenRepository) Save(ctx context.Context, userID int64, tokenHash string, duration time.Duration) error {
	sql := "INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(ctx, sql, userID, tokenHash, time.Now().Add(duration))
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
func (r *PostgresPasswordResetTokenRepository) Delete(ctx context.Context, tokenID int64) error {
	sql := "DELETE FROM password_reset_tokens WHERE id = $1"
	_, err := r.db.Exec(ctx, sql, tokenID)
	return err
}
