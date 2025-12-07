package usecase

import (
	"context"
	"log/slog"
	"strings"
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

	// Check if a user exists and verified
	exist, err := uc.userRepository.IsVerifiedUserExists(ctx, email)

	if err != nil {
		return err
	}

	if !exist {
		// Don't reveal to a client if user doesn't exist to prevent enumeration attack
		uc.logger.Warn("login OTP request for user that doesn't exist")
		return nil
	}

	// Generate code and hash
	code, err := uc.loginOTPRepository.GenerateCode(6)
	if err != nil {
		return err
	}

	codeHash := uc.loginOTPRepository.HashCode(code)

	// Save
	err = uc.loginOTPRepository.Save(ctx, email, codeHash)

	if err != nil {
		return err
	}

	// Dispatch task to send the OTP email
	return uc.taskDistributor.DistributeTaskSendEmailLoginOTP(ctx, email, code)
}
