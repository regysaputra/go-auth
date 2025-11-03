package usecase

import (
	"auth/internal/domain"
	"context"
)

// UserRepository represents the user repository interface
type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindById(ctx context.Context, id int64) (*domain.User, error)
	IsVerifiedUserExists(ctx context.Context, email string) (bool, error)
	SetVerified(ctx context.Context, userID int64) error
	UpdatePassword(ctx context.Context, userID int64, newPassword string) error
}
