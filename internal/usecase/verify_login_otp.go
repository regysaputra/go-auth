package usecase

import "context"

type VerifyLoginOTPUseCase struct {
	loginOTPRepository LoginOTPRepository
	userRepository     UserRepository
	loginUseCase       *LoginUserUseCase
}

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

func (uc *VerifyLoginOTPUseCase) Execute(ctx context.Context, code string) (*LoginToken, error) {
	// Validate code
	hashCode := uc.loginOTPRepository.Hash(code)
	err := uc.loginOTPRepository.IsCodeExist(ctx, hashCode)
	if err != nil {
		return nil, err
	}

	// Delete from db if code is valid
	err = uc.loginOTPRepository.Delete(ctx, hashCode)
	if err != nil {
		return nil, err
	}

	// Generate login token
	return uc.loginUseCase.GenerateToken(ctx, 123, false)
}
