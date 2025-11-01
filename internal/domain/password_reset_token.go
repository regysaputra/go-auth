package domain

import "time"

type PasswordResetToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}
