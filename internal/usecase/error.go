package usecase

import "errors"

// Pre-defined errors for specific business rule violations.
var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidToken            = errors.New("invalid token")
	ErrEmailExists             = errors.New("user with this email already exists")
	ErrInvalidInput            = errors.New("invalid input")
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrInvalidVerificationCode = errors.New("invalid or expired verification code")
	ErrEmptyName               = errors.New("name field is required")
	ErrEmptyEmail              = errors.New("email field is required")
	ErrEmptyPassword           = errors.New("password field is required")
	ErrPasswordTooShort        = errors.New("password must be at least 8 characters long")
	ErrInternalServer          = errors.New("internal server error")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserUnauthorized        = errors.New("user is unauthorized")
)
