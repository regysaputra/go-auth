package usecase

import (
	"auth/internal/domain"
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type RegisterUserUseCase struct {
	userRepository                   UserRepository
	sendEmailVerificationLinkUseCase *SendEmailVerificationLinkUseCase
}

func NewRegisterUserUseCase(
	userRepository UserRepository,
	sendEmailVerificationLinkUC *SendEmailVerificationLinkUseCase,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepository:                   userRepository,
		sendEmailVerificationLinkUseCase: sendEmailVerificationLinkUC,
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, name, email, password string) (*domain.User, error) {
	// Add validation for empty fields
	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyName
	}
	if strings.TrimSpace(email) == "" {
		return nil, ErrEmptyEmail
	}
	if strings.TrimSpace(password) == "" {
		return nil, ErrEmptyPassword
	}

	// email validation (domain level)
	userEmail := &domain.User{Email: email}

	if err := userEmail.Validate(); err != nil {
		return nil, ErrInvalidEmail
	}

	// password validation
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// Check if user already exist
	_, err := uc.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if err.Error() != "no rows in result set" {
			return nil, err
		}
	}

	if err == nil {
		return nil, ErrEmailExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create the domain user
	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := uc.userRepository.Save(ctx, user); err != nil {
		return nil, err
	}

	err = uc.sendEmailVerificationLinkUseCase.Execute(ctx, user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
