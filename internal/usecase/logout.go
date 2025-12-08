package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// LogoutUseCase represents the logout use case
type LogoutUseCase struct {
	refreshTokenRepository RefreshTokenRepository
	tokenRepository        TokenRepository
}

// NewLogoutUseCase creates a new logout use case
func NewLogoutUseCase(
	refreshTokenRepository RefreshTokenRepository,
	tokenRepository TokenRepository,
) *LogoutUseCase {
	return &LogoutUseCase{
		refreshTokenRepository: refreshTokenRepository,
		tokenRepository:        tokenRepository,
	}
}

// Execute performs the logout use case
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	// Field validation
	validationErrors := NewValidationErrors()
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		validationErrors.Add("refresh_token", "refresh_token field is required")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	// Check if a token is existing on a database
	refreshTokenHash := uc.tokenRepository.HashToken(refreshToken)
	oldRefreshToken, err := uc.refreshTokenRepository.FindByToken(ctx, refreshTokenHash)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}

		return err
	}

	// Revoke token
	err = uc.refreshTokenRepository.Revoke(ctx, oldRefreshToken.ID)
	if err != nil {
		return err
	}

	return nil
}
