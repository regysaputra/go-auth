package usecase

import (
	"context"
	"database/sql"
	"errors"
)

// VerifyEmailUseCase Handle the logic for verifying a user email with a token
type VerifyEmailUseCase struct {
	userRepository              UserRepository
	verificationTokenRepository VerificationTokenRepository
	loginUseCase                *LoginUserUseCase
}

// NewVerifyEmailUseCase creates a new VerifyEmailUseCase object
func NewVerifyEmailUseCase(userRepository UserRepository, verificationTokenRepository VerificationTokenRepository, loginUseCase *LoginUserUseCase) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{
		userRepository:              userRepository,
		verificationTokenRepository: verificationTokenRepository,
		loginUseCase:                loginUseCase,
	}
}

// Execute validates a token, marks the user as verified, and logs them in
func (uc *VerifyEmailUseCase) Execute(ctx context.Context, rawToken string) (*LoginToken, error) {
	if rawToken == "" {
		return nil, ErrInvalidToken
	}

	// Find the verification token
	token, err := uc.verificationTokenRepository.FindByToken(ctx, rawToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidToken
		}

		return nil, err
	}

	// Delete the token immediately so it can't be reused
	if err := uc.verificationTokenRepository.Delete(ctx, token.ID); err != nil {
		return nil, err
	}

	// Mark the user as verified
	if err := uc.userRepository.SetVerified(ctx, token.UserID); err != nil {
		return nil, err
	}

	// Log the user in by generating a JWT and a new remember token
	// A long-lived remember token is created by default upon verification
	loginToken, err := uc.loginUseCase.GenerateToken(ctx, token.UserID, true)
	if err != nil {
		return nil, err
	}

	return &LoginToken{
		AccessToken:   loginToken.AccessToken,
		RememberToken: loginToken.RememberToken,
	}, nil
}
