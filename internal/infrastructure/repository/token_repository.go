package repository

import (
	"auth/internal/usecase"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenRepository represents the token repository object
type TokenRepository struct {
	secretKey string
}

// NewTokenRepository creates a new JWT auth repository object
func NewTokenRepository(secretKey string) usecase.TokenRepository {
	return &TokenRepository{
		secretKey: secretKey,
	}
}

// GenerateJWTToken generates a JWT token
func (r *TokenRepository) GenerateJWTToken(subject any, duration time.Duration) (string, error) {
	// Create the token claims
	claims := jwt.MapClaims{
		"sub": subject,                         // Subject
		"iat": time.Now().Unix(),               // Issued At
		"exp": time.Now().Add(duration).Unix(), // Expiration Time
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(r.secretKey))
}

// GenerateOpaqueToken generates random string
func (r *TokenRepository) GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	token := base64.RawURLEncoding.EncodeToString(b)

	return token, nil
}

// HashToken hash the token
func (r *TokenRepository) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
