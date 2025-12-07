package usecase

import (
	"auth/internal/domain"
	"context"

	"github.com/google/uuid"
)

// EmailVerificationCodeRepository represents the email verification code repository interface
type EmailVerificationCodeRepository interface {
	GenerateCode(length int) (string, error)
	Save(ctx context.Context, email string, codeHash string) error
	HashCode(code string) string
	FindByEmail(ctx context.Context, email string) (*domain.EmailVerificationCode, error)
	FindByCode(ctx context.Context, code string) (*domain.EmailVerificationCode, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
