package usecase

import (
	"context"
	"database/sql"
	"errors"
)

// VerifyCodeUseCase represents the use case for verifying the code
type VerifyCodeUseCase struct {
	emailVerificationCodeRepository EmailVerificationCodeRepository
	tokenGenerator                  TokenGenerator
}

// NewVerifyCodeUseCase creates a new VerifyCodeUseCase object
func NewVerifyCodeUseCase(
	emailVerificationCodeRepository EmailVerificationCodeRepository,
	tokenGenerator TokenGenerator,
) *VerifyCodeUseCase {
	return &VerifyCodeUseCase{
		emailVerificationCodeRepository: emailVerificationCodeRepository,
		tokenGenerator:                  tokenGenerator,
	}
}

// Execute executes the use case
func (uc *VerifyCodeUseCase) Execute(ctx context.Context, code string) (string, error) {
	// Hash the code
	hashCode := uc.emailVerificationCodeRepository.Hash(code)
	verification, err := uc.emailVerificationCodeRepository.FindByCode(ctx, hashCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidVerificationCode
		}

		return "", err
	}

	token, err := uc.tokenGenerator.GenerateToken(verification.Email, "verification_token")
	if err != nil {
		return "", err
	}

	return token, nil
}
