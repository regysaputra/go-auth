package usecase

import (
	"auth/internal/domain"
	"context"
	"time"
)

type VerificationTokenRepository interface {
	Generate() (string, error)
	Hash(token string) string
	Save(ctx context.Context, userID int64, tokenHash string, duration time.Duration) error
	FindByToken(ctx context.Context, rawToken string) (*domain.VerificationToken, error)
	Delete(ctx context.Context, tokenID int64) error
}
