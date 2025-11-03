package usecase

import (
	"context"
	"time"
)

// LoginOTPRepository interface
type LoginOTPRepository interface {
	Generate(length int) (string, error)
	Hash(code string) string
	Save(ctx context.Context, email string, codeHash string, duration time.Duration) error
	IsCodeExist(ctx context.Context, codeHash string) error
	Delete(ctx context.Context, email string) error
}
