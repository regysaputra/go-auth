package usecase

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ResetPasswordUseCase represents the reset password use case object
type ResetPasswordUseCase struct {
	userRepository               UserRepository
	tokenRepository              TokenRepository
	passwordResetTokenRepository PasswordResetTokenRepository
}

// NewResetPasswordUseCase creates a new reset password use case object
func NewResetPasswordUseCase(userRepository UserRepository, tokenRepository TokenRepository, passwordResetTokenRepository PasswordResetTokenRepository) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepository:               userRepository,
		tokenRepository:              tokenRepository,
		passwordResetTokenRepository: passwordResetTokenRepository,
	}
}

// Execute executes the reset password use case
func (uc *ResetPasswordUseCase) Execute(ctx context.Context, token, newPassword string) error {
	// Field validation
	validationErrors := NewValidationErrors()
	if newPassword == "" {
		validationErrors.Add("password", "password field is required")
	} else if len(newPassword) < 8 {
		validationErrors.Add("password", "password must be at least 8 characters")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	// Check if hashed token is exist in database
	tokenHash := uc.tokenRepository.HashToken(token)
	resetToken, err := uc.passwordResetTokenRepository.FindByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}

		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	err = uc.userRepository.UpdatePassword(ctx, resetToken.UserID, string(hashedPassword))
	if err != nil {
		return err
	}

	// Invalidate the token after use
	err = uc.passwordResetTokenRepository.Delete(ctx, resetToken.ID)
	if err != nil {
		return err
	}

	return nil
}
