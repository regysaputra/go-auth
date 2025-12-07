package usecase

import (
	"auth/internal/domain"
	"context"

	"github.com/google/uuid"
)

// PasswordResetTokenRepository represents the password reset token repository interface
type PasswordResetTokenRepository interface {
	Save(ctx context.Context, userID int64, tokenHash string) error
	FindByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
