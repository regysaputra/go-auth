package usecase

import (
	"auth/internal/domain"
	"context"
)

// LoginOTPRepository interface
type LoginOTPRepository interface {
	GenerateCode(length int) (string, error)
	HashCode(code string) string
	Save(ctx context.Context, email string, codeHash string) error
	IsCodeExist(ctx context.Context, codeHash string) error
	FindByCode(ctx context.Context, codeHash string) (*domain.LoginOTP, error)
	Delete(ctx context.Context, email string) error
}
