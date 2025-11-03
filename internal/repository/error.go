package repository

import "errors"

var (
	// ErrNotFound is returned when a specific record is not found in the database.
	ErrNotFound = errors.New("record not found")
)

// ErrDuplicateEmail is returned when a user with the same email already exists.
type ErrDuplicateEmail struct{}

func (err *ErrDuplicateEmail) Error() string {
	return "email already exists"
}
