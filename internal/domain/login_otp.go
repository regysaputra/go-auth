package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LoginOTP represents a login OTP
type LoginOTP struct {
	ID        uuid.UUID
	Email     string
	CodeHash  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Validate validates the LoginOTP
func (obj *LoginOTP) Validate() error {
	if !strings.Contains(obj.Email, "@") {
		return errors.New("invalid email format")
	}

	return nil
}
