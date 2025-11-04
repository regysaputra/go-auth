package usecase

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"errors"
)

// GetUserProfileUseCase represents the GetUserProfile use case object
type GetUserProfileUseCase struct {
	UserRepository UserRepository
}

// NewGetUserProfileUseCase creates a new GetUserProfile use case object
func NewGetUserProfileUseCase(userRepository UserRepository) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{userRepository}
}

// Execute executes the GetUserProfile use case
func (uc *GetUserProfileUseCase) Execute(ctx context.Context, userID int64) (*domain.User, error) {
	// Find user by their id
	user, err := uc.UserRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}
