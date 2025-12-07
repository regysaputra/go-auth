package usecase

import (
	"auth/internal/domain"
	"context"

	"github.com/google/uuid"
)

// VerificationTokenRepository represents the verification token repository interface
type VerificationTokenRepository interface {
	Save(ctx context.Context, userID int64, tokenHash string) error
	FindByToken(ctx context.Context, tokenHash string) (*domain.VerificationToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
