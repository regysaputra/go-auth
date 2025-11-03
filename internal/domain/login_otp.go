package domain

import (
	"errors"
	"strings"
	"time"
)

// LoginOTP represents a login OTP
type LoginOTP struct {
	ID        int64
	Email     string
	CodeHash  string
	ExpiresAt time.Time
}

// Validate validates the LoginOTP
func (obj *LoginOTP) Validate() error {
	if !strings.Contains(obj.Email, "@") {
		return errors.New("invalid email format")
	}

	return nil
}
