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

type PostgresVerificationTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresVerificationTokenRepository(db *pgxpool.Pool) *PostgresVerificationTokenRepository {
	return &PostgresVerificationTokenRepository{
		db: db,
	}
}

func (r *PostgresVerificationTokenRepository) Generate() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func (r *PostgresVerificationTokenRepository) Hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

func (r *PostgresVerificationTokenRepository) Save(ctx context.Context, userID int64, tokenHash string, duration time.Duration) error {
	sql := `INSERT INTO verification_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	expiresAt := time.Now().Add(duration)
	_, err := r.db.Exec(ctx, sql, userID, tokenHash, expiresAt)
	return err
}

func (r *PostgresVerificationTokenRepository) FindByToken(ctx context.Context, rawToken string) (*domain.VerificationToken, error) {
	sql := `SELECT * FROM verification_tokens WHERE token_hash = $1 AND expires_at > NOW()`
	tokenHash := r.Hash(rawToken)
	var token domain.VerificationToken
	row := r.db.QueryRow(ctx, sql, tokenHash)
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *PostgresVerificationTokenRepository) Delete(ctx context.Context, tokenId int64) error {
	sql := `DELETE FROM verification_tokens WHERE id = $1`
	_, err := r.db.Exec(ctx, sql, tokenId)
	return err
}
