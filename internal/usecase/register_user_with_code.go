package usecase

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUserWithCodeUseCase represents the RegisterUserWithCode use case object
type RegisterUserWithCodeUseCase struct {
	userRepository    UserRepository
	verifyCodeUseCase *VerifyCodeUseCase
	loginUseCase      *LoginUserUseCase
}

// Claims represents the JWT claims
type Claims struct {
	Subject string `json:"sub"`
	Issuer  string `json:"iss"`
	Expires int64  `json:"exp"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

// NewRegisterUserWithCodeUseCase creates a new RegisterUserWithCodeUseCase object
func NewRegisterUserWithCodeUseCase(
	userRepository UserRepository,
	verifyCodeUseCase *VerifyCodeUseCase,
	loginUseCase *LoginUserUseCase,
) *RegisterUserWithCodeUseCase {
	return &RegisterUserWithCodeUseCase{
		userRepository:    userRepository,
		verifyCodeUseCase: verifyCodeUseCase,
		loginUseCase:      loginUseCase,
	}
}

// Execute executes the RegisterUserWithCode use case
func (uc *RegisterUserWithCodeUseCase) Execute(ctx context.Context, verificationToken string, name string, password string) (*LoginToken, error) {
	// Decode & verify JWT signature
	token, err := jwt.ParseWithClaims(verificationToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check expiration time
	if claims.Expires > time.Now().Unix() {
		return nil, ErrInvalidToken
	}

	// Check purpose
	if claims.Purpose != "verificationToken" {
		return nil, ErrInvalidToken
	}

	// Extract email
	email := claims.Subject
	if strings.TrimSpace(email) == "" {
		return nil, ErrInvalidToken
	}

	// Add validation for empty fields
	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyName
	}
	if strings.TrimSpace(password) == "" {
		return nil, ErrEmptyPassword
	}

	// password validation
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	// Check if a user already exists
	_, err = uc.userRepository.IsVerifiedUserExists(ctx, email)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
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

	// Generate login token
	return uc.loginUseCase.GenerateToken(ctx, user.ID, false)
}
