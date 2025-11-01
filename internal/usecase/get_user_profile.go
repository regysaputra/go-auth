package usecase

import (
	"auth/internal/domain"
	"context"
)

type GetUserProfileUseCase struct {
	UserRepository UserRepository
}

func NewGetUserProfileUseCase(userRepository UserRepository) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{userRepository}
}

func (uc *GetUserProfileUseCase) Execute(ctx context.Context, userID int64) (*domain.User, error) {
	// Find user by their id
	user, err := uc.UserRepository.FindById(ctx, userID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return user, nil
}
