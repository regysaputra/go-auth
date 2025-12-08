package usecase

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
)

// RequestPasswordResetUseCase represents the use case for requesting password reset
type RequestPasswordResetUseCase struct {
	logger                       *slog.Logger
	userRepository               UserRepository
	passwordResetTokenRepository PasswordResetTokenRepository
	tokenRepository              TokenRepository
	taskDistributor              TaskDistributor
}

// NewRequestPasswordResetUseCase creates a new RequestPasswordResetUseCase object
func NewRequestPasswordResetUseCase(
	logger *slog.Logger,
	userRepository UserRepository,
	passwordResetTokenRepository PasswordResetTokenRepository,
	tokenRepository TokenRepository,
	taskDistributor TaskDistributor,
) *RequestPasswordResetUseCase {
	return &RequestPasswordResetUseCase{
		logger:                       logger,
		userRepository:               userRepository,
		passwordResetTokenRepository: passwordResetTokenRepository,
		tokenRepository:              tokenRepository,
		taskDistributor:              taskDistributor,
	}
}

// Execute executes the use case
func (uc *RequestPasswordResetUseCase) Execute(ctx context.Context, email string) error {
	// Field validation
	validationErrors := NewValidationErrors()
	email = strings.TrimSpace(email)
	if email == "" {
		validationErrors.Add("email", "email field is required")
	} else if !strings.Contains(email, "@") {
		validationErrors.Add("email", "email must be a valid email address")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	// Check if a user exists
	user, err := uc.userRepository.FindByEmail(ctx, email)

	// don't throw error if the user doesn't exist or unverified to prevent email enumeration attack
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			uc.logger.Warn("Password reset requested for non-existent user", "email", email)
			return nil
		}

		uc.logger.Error("internal server error", "error", err)
		return nil
	}

	if !user.Verified {
		uc.logger.Warn("Password reset requested for non-verified user", "email", email)
		return nil
	}

	// Generate token
	token, err := uc.tokenRepository.GenerateOpaqueToken()
	if err != nil {
		return err
	}

	// Hash token
	tokenHash := uc.tokenRepository.HashToken(token)

	// Save hash token to a database
	err = uc.passwordResetTokenRepository.Save(ctx, user.ID, tokenHash)
	if err != nil {
		return err
	}

	err = uc.taskDistributor.DistributeTaskSendEmailPasswordResetLink(ctx, user.Email, token)

	if err != nil {
		return err
	}

	return nil
}
