package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// RefreshTokenUseCase handle the logic for exchanging a remember token for a new JWT
type RefreshTokenUseCase struct {
	userRepository          UserRepository
	rememberTokenRepository RememberTokenRepository
	tokenGenerator          TokenGenerator
}

// RefreshResult Hold the output of a successful token refresh
type RefreshResult struct {
	NewJWT           string
	NewRememberToken string
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase object
func NewRefreshTokenUseCase(
	userRepository UserRepository,
	rememberTokenRepository RememberTokenRepository,
	tokenGenerator TokenGenerator,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepository:          userRepository,
		rememberTokenRepository: rememberTokenRepository,
		tokenGenerator:          tokenGenerator,
	}
}

// Execute validates a remember token, performs secure token rotation and issues a new JWT
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, rawRememberToken string) (*RefreshResult, error) {
	if rawRememberToken == "" {
		return nil, ErrInvalidCredentials
	}
	// Find the remember token in the database and check if it's expired
	hashToken := uc.rememberTokenRepository.Hash(rawRememberToken)
	oldToken, err := uc.rememberTokenRepository.FindByToken(ctx, hashToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}

		return nil, err
	}

	// Immediately delete the used token to prevent replay attacks
	err = uc.rememberTokenRepository.Delete(ctx, oldToken.ID)
	if err != nil {
		return nil, err
	}

	// Verify the user associated with the token still exists
	_, err = uc.userRepository.FindByID(ctx, oldToken.UserID)

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Issue a new JWT for the user
	newJWT, err := uc.tokenGenerator.GenerateToken(oldToken.UserID, "refresh_token")

	if err != nil {
		return nil, err
	}

	// Issue a new remember token
	newRememberToken, err := uc.rememberTokenRepository.Generate()
	if err != nil {
		return nil, err
	}

	newHash := uc.rememberTokenRepository.Hash(newRememberToken)

	// Use the same duration as the login use case
	rememberTokenDuration := time.Hour * 24 * 30
	if err := uc.rememberTokenRepository.Save(ctx, oldToken.UserID, newHash, rememberTokenDuration); err != nil {
		return nil, err
	}

	result := &RefreshResult{
		NewJWT:           newJWT,
		NewRememberToken: newRememberToken,
	}

	return result, nil
}
