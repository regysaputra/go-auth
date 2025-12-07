package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// VerifyCodeUseCase represents the use case for verifying the code
type VerifyCodeUseCase struct {
	emailVerificationCodeRepository EmailVerificationCodeRepository
	tokenGenerator                  TokenRepository
	registrationTokenTTL            time.Duration
}

// NewVerifyCodeUseCase creates a new VerifyCodeUseCase object
func NewVerifyCodeUseCase(
	emailVerificationCodeRepository EmailVerificationCodeRepository,
	tokenGenerator TokenRepository,
	registrationTokenTTL time.Duration,
) *VerifyCodeUseCase {
	return &VerifyCodeUseCase{
		emailVerificationCodeRepository: emailVerificationCodeRepository,
		tokenGenerator:                  tokenGenerator,
		registrationTokenTTL:            registrationTokenTTL,
	}
}

// Execute executes the use case
func (uc *VerifyCodeUseCase) Execute(ctx context.Context, code string) (string, error) {
	// Hash the code
	hashCode := uc.emailVerificationCodeRepository.HashCode(code)
	verification, err := uc.emailVerificationCodeRepository.FindByCode(ctx, hashCode)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidVerificationCode
		}

		return "", err
	}

	token, err := uc.tokenGenerator.GenerateJWTToken(verification.Email, uc.registrationTokenTTL)
	if err != nil {
		return "", err
	}

	return token, nil
}
