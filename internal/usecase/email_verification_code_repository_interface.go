package usecase

import (
	"auth/internal/domain"
	"context"
	"time"
)

type EmailVerificationCodeRepository interface {
	GenerateCode(length int) (string, error)
	Save(ctx context.Context, email string, codeHash string, duration time.Duration) error
	Hash(code string) string
	FindByEmail(ctx context.Context, email string) (*domain.EmailVerificationCode, error)
	FindByCode(ctx context.Context, code string) (*domain.EmailVerificationCode, error)
	Delete(ctx context.Context, id int64) error
}
