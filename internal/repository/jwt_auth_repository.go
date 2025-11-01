package repository

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthRepository struct {
	secretKey string
}

func NewJWTAuthRepository() *JWTAuthRepository {
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		// Provide a default for local development. In production, this MUST be set.
		secret = "a-very-secure-and-long-secret-key-for-dev"
	}

	return &JWTAuthRepository{secretKey: secret}
}

func (r *JWTAuthRepository) GenerateToken(subject any, purpose string) (string, error) {
	// Create the token claims
	claims := jwt.MapClaims{
		"sub":     subject,                               // Subject
		"iat":     time.Now().Unix(),                     // Issued At
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expiration Time
		"purpose": purpose,
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(r.secretKey))
}
