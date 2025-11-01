package repository

import "errors"

var (
	// ErrNotFound is returned when a specific record is not found in the database.
	ErrNotFound = errors.New("record not found")
)

type ErrDuplicateEmail struct{}

func (err *ErrDuplicateEmail) Error() string {
	return "email already exists"
}
