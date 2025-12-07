package usecase

import (
	"context"
	"database/sql"
	"errors"
	"net"
)

// VerifyLoginOTPUseCase represents the use case for verifying login otp
type VerifyLoginOTPUseCase struct {
	loginOTPRepository LoginOTPRepository
	userRepository     UserRepository
	loginUseCase       *LoginUserUseCase
}

// NewVerifyLoginOTPUseCase creates a new VerifyLoginOTPUseCase object
func NewVerifyLoginOTPUseCase(
	loginOTPRepository LoginOTPRepository,
	userRepository UserRepository,
	loginUseCase *LoginUserUseCase,
) *VerifyLoginOTPUseCase {
	return &VerifyLoginOTPUseCase{
		loginOTPRepository: loginOTPRepository,
		userRepository:     userRepository,
		loginUseCase:       loginUseCase,
	}
}

// Execute verify login otp
func (uc *VerifyLoginOTPUseCase) Execute(ctx context.Context, code string, userAgent string, ipAddress net.IP) (*LoginResult, error) {
	// Hash code
	hashCode := uc.loginOTPRepository.HashCode(code)

	// Find login otp by code hash
	loginOTP, err := uc.loginOTPRepository.FindByCode(ctx, hashCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	// Delete from db if code is valid
	err = uc.loginOTPRepository.Delete(ctx, hashCode)
	if err != nil {
		return nil, err
	}

	// Find user by email
	user, err := uc.userRepository.FindByEmail(ctx, loginOTP.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	// Generate login token
	return uc.loginUseCase.GenerateLoginToken(ctx, user, false, userAgent, ipAddress)
}
