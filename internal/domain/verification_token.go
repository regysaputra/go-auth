package domain

import "time"

// VerificationToken represents a verification token
type VerificationToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}
