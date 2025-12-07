package usecase

import (
	"context"
	"strings"
)

// RequestVerificationCodeUseCase represents the request verification code use case object
type RequestVerificationCodeUseCase struct {
	userRepository                  UserRepository
	emailVerificationCodeRepository EmailVerificationCodeRepository
	taskDistributor                 TaskDistributor
}

// NewRequestVerificationCodeUseCase creates a new request verification code use case object
func NewRequestVerificationCodeUseCase(
	emailVerificationCodeRepository EmailVerificationCodeRepository,
	userRepository UserRepository,
	taskDistributor TaskDistributor,
) *RequestVerificationCodeUseCase {
	return &RequestVerificationCodeUseCase{
		emailVerificationCodeRepository: emailVerificationCodeRepository,
		userRepository:                  userRepository,
		taskDistributor:                 taskDistributor,
	}
}

// Execute executes the request verification code use case
func (uc *RequestVerificationCodeUseCase) Execute(ctx context.Context, email string) error {
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

	// Check if verified user exists with given email
	exist, err := uc.userRepository.IsVerifiedUserExists(ctx, email)

	if err != nil {
		return err
	}

	if exist {
		return ErrEmailExists
	}

	// Generate and hash code
	code, err := uc.emailVerificationCodeRepository.GenerateCode(6)
	if err != nil {
		return err
	}

	hashCode := uc.emailVerificationCodeRepository.HashCode(code)

	// Save to db
	err = uc.emailVerificationCodeRepository.Save(ctx, email, hashCode)
	if err != nil {
		return err
	}

	// Dispatch background task to send email
	err = uc.taskDistributor.DistributeTaskSendEmailVerificationCode(ctx, email, code)

	if err != nil {
		return err
	}

	return nil
}
