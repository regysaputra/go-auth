package usecase

import (
	"context"
	"database/sql"
	"errors"
	"net"
)

// VerifyTokenUseCase Handle the logic for verifying a user email with a token
type VerifyTokenUseCase struct {
	userRepository              UserRepository
	verificationTokenRepository VerificationTokenRepository
	loginUseCase                *LoginUserUseCase
}

// NewVerifyTokenUseCase creates a new VerifyTokenUseCase object
func NewVerifyTokenUseCase(userRepository UserRepository, verificationTokenRepository VerificationTokenRepository, loginUseCase *LoginUserUseCase) *VerifyTokenUseCase {
	return &VerifyTokenUseCase{
		userRepository:              userRepository,
		verificationTokenRepository: verificationTokenRepository,
		loginUseCase:                loginUseCase,
	}
}

// Execute validates a token, marks the user as verified, and logs them in
func (uc *VerifyTokenUseCase) Execute(ctx context.Context, rawToken string, userAgent string, ipAddress net.IP) (*LoginResult, error) {
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

	// Find the user by their ID
	user, err := uc.userRepository.FindByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	// Log the user in by generating a JWT and a new remember token
	// A long-lived remember token is created by default upon verification
	return uc.loginUseCase.GenerateLoginToken(ctx, user, false, userAgent, ipAddress)
}
