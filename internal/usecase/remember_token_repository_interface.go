package usecase

import (
	"auth/internal/domain"
	"context"
	"time"
)

type RememberTokenRepository interface {
	// Generate creates a new secure token string.
	Generate() (string, error)
	// Hash hashes a raw token string using SHA-256.
	Hash(token string) string
	// Save stores a new remember token in the database.
	Save(ctx context.Context, userID int64, tokenHash string, duration time.Duration) error
	// FindByToken hashes the provided raw token and finds the matching record.
	FindByToken(ctx context.Context, hashToken string) (*domain.RememberToken, error)
	// Delete removes a token by its ID.
	Delete(ctx context.Context, tokenID int64) error
}
