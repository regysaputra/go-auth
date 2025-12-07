package domain

import (
	"time"

	"github.com/google/uuid"
)

// EmailVerificationCode to hold data of email verification code
type EmailVerificationCode struct {
	ID        uuid.UUID
	Email     string
	CodeHash  string
	ExpiresAt time.Time
	CreatedAt time.Time
}
