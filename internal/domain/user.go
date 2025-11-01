package domain

import (
	"errors"
	"strings"
)

// User represent a user in the system
// @Description User information
// @Description with id, name, email, and password
type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Verified bool   `json:"verified"`
}

func (u *User) Validate() error {
	if !strings.Contains(u.Email, "@") {
		return errors.New("invalid email format")
	}

	return nil
}
