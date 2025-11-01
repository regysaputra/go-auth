package usecase

import (
	"context"
)

type VerifyCodeUseCase struct {
	emailVerificationCodeRepository EmailVerificationCodeRepository
	tokenGenerator                  TokenGenerator
}

func NewVerifyCodeUseCase(
	emailVerificationCodeRepository EmailVerificationCodeRepository,
	tokenGenerator TokenGenerator,
) *VerifyCodeUseCase {
	return &VerifyCodeUseCase{
		emailVerificationCodeRepository: emailVerificationCodeRepository,
		tokenGenerator:                  tokenGenerator,
	}
}

func (uc *VerifyCodeUseCase) Execute(ctx context.Context, code string) (string, error) {
	// Hash the code
	hashCode := uc.emailVerificationCodeRepository.Hash(code)
	verification, err := uc.emailVerificationCodeRepository.FindByCode(ctx, hashCode)
	if err != nil {
		if err.Error() == "no rows in result set" {
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
