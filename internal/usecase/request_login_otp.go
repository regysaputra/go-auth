package usecase

import (
	"auth/internal/domain"
	"context"
	"log/slog"
	"strings"
	"time"
)

// RequestLoginOTPUseCase represents the request login OTP use case object
type RequestLoginOTPUseCase struct {
	logger             *slog.Logger
	loginOTPRepository LoginOTPRepository
	userRepository     UserRepository
	taskDistributor    TaskDistributor
}

// NewRequestLoginOTPUseCase creates a new request login OTP use case object
func NewRequestLoginOTPUseCase(
	logger *slog.Logger,
	loginOTPRepository LoginOTPRepository,
	userRepository UserRepository,
	taskDistributor TaskDistributor,
) *RequestLoginOTPUseCase {
	return &RequestLoginOTPUseCase{
		logger:             logger,
		loginOTPRepository: loginOTPRepository,
		userRepository:     userRepository,
		taskDistributor:    taskDistributor,
	}
}

// Execute executes the request login OTP use case
func (uc *RequestLoginOTPUseCase) Execute(ctx context.Context, email string) error {
	// Validate input
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmptyEmail
	}

	if err := (&domain.EmailVerificationCode{Email: email}).Validate(); err != nil {
		return ErrInvalidEmail
	}

	// Check if user exist and verified
	exist, err := uc.userRepository.IsVerifiedUserExists(ctx, email)

	if err != nil {
		return err
	}

	if !exist {
		// Don't reveal to client if user doesn't exist to prevent enumeration attack
		uc.logger.Warn("login OTP request for user that doesn't exist")
		return nil
	}

	// Generate code and hash
	code, err := uc.loginOTPRepository.Generate(6)
	if err != nil {
		return err
	}

	codeHash := uc.loginOTPRepository.Hash(code)

	// Save
	err = uc.loginOTPRepository.Save(ctx, email, codeHash, 5*time.Minute)
	if err != nil {
		return err
	}

	// Dispatch task to send the OTP email
	return uc.taskDistributor.DistributeTaskSendEmailLoginOTP(ctx, email, code)
}
