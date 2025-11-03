package usecase

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// ResetPasswordUseCase represents the reset password use case object
type ResetPasswordUseCase struct {
	userRepository  UserRepository
	tokenRepository PasswordResetTokenRepository
}

// NewResetPasswordUseCase creates a new reset password use case object
func NewResetPasswordUseCase(userRepository UserRepository, tokenRepository PasswordResetTokenRepository) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
	}
}

// Execute executes the reset password use case
func (uc *ResetPasswordUseCase) Execute(ctx context.Context, token, newPassword string) error {
	// Length of new password must be greater than 7
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	// Check if hashed token is exist in database
	resetToken, err := uc.tokenRepository.FindByToken(ctx, uc.tokenRepository.Hash(token))
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
	err = uc.tokenRepository.Delete(ctx, resetToken.ID)
	if err != nil {
		return err
	}

	return nil
}
