package domain

import (
	"time"

	"github.com/google/uuid"
)

// VerificationToken represents a verification token
type VerificationToken struct {
	ID        uuid.UUID
	UserID    int64
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
}
