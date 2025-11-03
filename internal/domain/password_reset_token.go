package domain

import "time"

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}
