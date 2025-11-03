package usecase

import (
	"auth/internal/domain"
	"context"
	"time"
)

// PasswordResetTokenRepository represents the password reset token repository interface
type PasswordResetTokenRepository interface {
	Generate() (string, error)
	Hash(token string) string
	Save(ctx context.Context, userID int64, tokenHash string, duration time.Duration) error
	FindByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error)
	Delete(ctx context.Context, tokenID int64) error
}
