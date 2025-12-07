package usecase

import "time"

// TokenRepository interface for token generation
type TokenRepository interface {
	GenerateJWTToken(subject any, duration time.Duration) (string, error)
	GenerateOpaqueToken() (string, error)
	HashToken(token string) string
}
