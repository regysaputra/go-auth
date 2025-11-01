package usecase

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//type LoginToken struct {
//	AccessToken   string
//	RememberToken string
//}

type LoginUserUseCase struct {
	userRepository     UserRepository
	tokenGenerator     TokenGenerator
	rememberRepository RememberTokenRepository
	rememberMeHours    time.Duration
}

type LoginToken struct {
	AccessToken   string
	RememberToken string
}

func NewLoginUserUseCase(userRepository UserRepository, tokenGenerator TokenGenerator, rememberRepository RememberTokenRepository) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepository:     userRepository,
		tokenGenerator:     tokenGenerator,
		rememberRepository: rememberRepository,
		rememberMeHours:    time.Hour * 24 * 30,
	}
}

// Execute authenticates a user by checking their credentials and then generates tokens for them
func (uc *LoginUserUseCase) Execute(ctx context.Context, email string, password string, rememberMe bool) (*LoginToken, error) {
	// find user by email
	user, err := uc.userRepository.FindByEmail(ctx, email)
	if err != nil {
		// Don't return user not found error to prevent email enumeration attack
		if err.Error() == "no rows in result set" {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	// compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return uc.GenerateToken(ctx, user.ID, rememberMe)
}

// GenerateToken Creates a new JWT and optionally a remember me token for a given user ID
// This method is separate from Execute so it can be called directly after other authentication flows, like email verification.
func (uc *LoginUserUseCase) GenerateToken(ctx context.Context, userID int64, rememberMe bool) (*LoginToken, error) {
	// generate access token for authenticated user
	token, err := uc.tokenGenerator.GenerateToken(userID, "access_token")

	if err != nil {
		return nil, err
	}

	result := &LoginToken{AccessToken: token}

	if rememberMe {
		rawToken, err := uc.rememberRepository.Generate()

		if err != nil {
			return nil, err
		}
		tokenHash := uc.rememberRepository.Hash(rawToken)
		err = uc.rememberRepository.Save(ctx, userID, tokenHash, uc.rememberMeHours)
		if err != nil {
			return nil, err
		}
		result.RememberToken = rawToken
	}
	fmt.Println("RESULT : ", result.RememberToken)
	return result, nil
}
