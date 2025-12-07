package domain

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    int64
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
}
