package usecase

import (
	"auth/internal/domain"
	"context"
	"strings"
	"time"
)

type RequestVerificationCodeUseCase struct {
	userRepository                  UserRepository
	emailVerificationCodeRepository EmailVerificationCodeRepository
	taskDistributor                 TaskDistributor
}

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

func (uc *RequestVerificationCodeUseCase) Execute(ctx context.Context, email string) error {
	// Email validation
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmptyEmail
	}

	userEmail := &domain.EmailVerificationCode{Email: email}
	err := userEmail.Validate()
	if err != nil {
		return ErrInvalidEmail
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

	hashCode := uc.emailVerificationCodeRepository.Hash(code)

	// Save to db
	err = uc.emailVerificationCodeRepository.Save(ctx, email, hashCode, 2*time.Minute)
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
