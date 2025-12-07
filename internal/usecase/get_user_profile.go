package usecase

import (
	"context"
	"database/sql"
	"errors"
)

// GetUserProfileResult represents the user profile response object
type GetUserProfileResult struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetUserProfileUseCase represents the GetUserProfile use case object
type GetUserProfileUseCase struct {
	UserRepository UserRepository
}

// NewGetUserProfileUseCase creates a new GetUserProfile use case object
func NewGetUserProfileUseCase(userRepository UserRepository) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{userRepository}
}

// Execute executes the GetUserProfile use case
func (uc *GetUserProfileUseCase) Execute(ctx context.Context, userID int64) (*GetUserProfileResult, error) {
	// Find user by their id
	user, err := uc.UserRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	response := &GetUserProfileResult{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return response, nil
}
