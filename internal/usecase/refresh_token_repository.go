package usecase

import (
	"auth/internal/domain"
	"context"

	"github.com/google/uuid"
)

// RefreshTokenRepository represents the refresh token repository interface.
type RefreshTokenRepository interface {
	// Save stores a new refresh token in the database.
	Save(ctx context.Context, refreshToken domain.RefreshToken) (uuid.UUID, error)
	// FindByToken hashes the provided raw token and finds the matching record.
	FindByToken(ctx context.Context, hashToken string) (*domain.RefreshToken, error)
	// Delete removes a token by its ID.
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeChain(ctx context.Context, id uuid.UUID) error
	Replace(ctx context.Context, oldID uuid.UUID, newID uuid.UUID) error
}
