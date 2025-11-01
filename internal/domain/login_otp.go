package domain

import (
	"errors"
	"strings"
	"time"
)

type LoginOTP struct {
	ID        int64
	Email     string
	CodeHash  string
	ExpiresAt time.Time
}

func (obj *LoginOTP) Validate() error {
	if !strings.Contains(obj.Email, "@") {
		return errors.New("invalid email format")
	}

	return nil
}
