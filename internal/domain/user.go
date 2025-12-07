package domain

import "time"

// User represent a user in the system
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
}
