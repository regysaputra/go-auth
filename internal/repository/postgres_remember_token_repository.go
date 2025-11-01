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

type PostgresRememberTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRememberTokenRepository(db *pgxpool.Pool) *PostgresRememberTokenRepository {
	return &PostgresRememberTokenRepository{db: db}
}

func (r *PostgresRememberTokenRepository) Generate() (string, error) {
	// Generate 32 bytes of random data
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to a URL-safe base64 string
	return base64.URLEncoding.EncodeToString(b), nil
}

func (r *PostgresRememberTokenRepository) Hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

func (r *PostgresRememberTokenRepository) Save(ctx context.Context, userID int64, tokenHash string, duration time.Duration) error {
	sql := "INSERT INTO remember_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)"
	expiresAt := time.Now().Add(duration)
	_, err := r.db.Exec(ctx, sql, userID, tokenHash, expiresAt)

	return err
}

func (r *PostgresRememberTokenRepository) FindByToken(ctx context.Context, hashToken string) (*domain.RememberToken, error) {
	sql := "SELECT * FROM remember_tokens WHERE token_hash = $1 AND expires_at > NOW()"

	row := r.db.QueryRow(ctx, sql, hashToken)
	var rememberToken domain.RememberToken
	err := row.Scan(&rememberToken.ID, &rememberToken.UserID, &rememberToken.TokenHash, &rememberToken.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return &rememberToken, nil
}

func (r *PostgresRememberTokenRepository) Delete(ctx context.Context, tokenID int64) error {
	sql := "DELETE FROM remember_tokens WHERE id = $1"
	_, err := r.db.Exec(ctx, sql, tokenID)
	return err
}
