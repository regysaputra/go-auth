package domain

import (
	"errors"
	"strings"
	"time"
)

type EmailVerificationCode struct {
	ID        int64
	Email     string
	CodeHash  string
	ExpiresAt time.Time
}

func (obj *EmailVerificationCode) Validate() error {
	if !strings.Contains(obj.Email, "@") {
		return errors.New("invalid email format")
	}

	return nil
}
