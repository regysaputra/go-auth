package usecase

import (
	"context"
	"log/slog"
	"time"
)

type RequestPasswordResetUseCase struct {
	logger          *slog.Logger
	userRepository  UserRepository
	tokenRepository PasswordResetTokenRepository
	taskDistributor TaskDistributor
}

func NewRequestPasswordResetUseCase(
	logger *slog.Logger,
	userRepository UserRepository,
	tokenRepository PasswordResetTokenRepository,
	taskDistributor TaskDistributor,
) *RequestPasswordResetUseCase {
	return &RequestPasswordResetUseCase{
		logger:          logger,
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		taskDistributor: taskDistributor,
	}
}

func (uc *RequestPasswordResetUseCase) Execute(ctx context.Context, email string) error {
	// Check if user exist
	user, err := uc.userRepository.FindByEmail(ctx, email)

	// don't throw error if user doesn't exist or unverified to prevent email enumeration attack
	if err != nil {
		if err.Error() == "no rows in result set" {
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
	token, err := uc.tokenRepository.Generate()
	if err != nil {
		return err
	}

	// Hash token
	tokenHash := uc.tokenRepository.Hash(token)

	// Save hash token to database
	err = uc.tokenRepository.Save(ctx, user.ID, tokenHash, time.Minute*15)
	if err != nil {
		return err
	}

	err = uc.taskDistributor.DistributeTaskSendEmailPasswordResetLink(ctx, user.Email, token)

	if err != nil {
		return err
	}

	return nil
}
