package usecase

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUserUseCase represents the register user use case
type RegisterUserUseCase struct {
	userRepository                   UserRepository
	sendEmailVerificationLinkUseCase *SendEmailVerificationLinkUseCase
}

// NewRegisterUserUseCase creates a new register user use case
func NewRegisterUserUseCase(
	userRepository UserRepository,
	sendEmailVerificationLinkUC *SendEmailVerificationLinkUseCase,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepository:                   userRepository,
		sendEmailVerificationLinkUseCase: sendEmailVerificationLinkUC,
	}
}

// Execute executes the register user use case
func (uc *RegisterUserUseCase) Execute(ctx context.Context, name, email, password string) error {
	// Field validation
	validationErrors := NewValidationErrors()
	email = strings.TrimSpace(email)
	if email == "" {
		validationErrors.Add("email", "email field is required")
	} else if !strings.Contains(email, "@") {
		validationErrors.Add("email", "email must be a valid email address")
	}

	if password == "" {
		validationErrors.Add("password", "password field is required")
	} else if len(password) < 8 {
		validationErrors.Add("password", "password must be at least 8 characters")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	// Check if a user already exists
	_, err := uc.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	if err == nil {
		return ErrEmailExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create the domain user
	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Verified: false,
	}

	if err := uc.userRepository.Save(ctx, user); err != nil {
		return err
	}

	err = uc.sendEmailVerificationLinkUseCase.Execute(ctx, user.ID, user.Email)
	if err != nil {
		return err
	}

	return nil
}
