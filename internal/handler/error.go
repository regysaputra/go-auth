package handler

import "errors"

var (
	// ErrMalformedAuthHeader is returned when the auth header is malformed
	ErrMalformedAuthHeader = errors.New("malformed auth header")

	// ErrMethodNotAllowed is returned when the method is not allowed
	ErrMethodNotAllowed = errors.New("method not allowed")

	// ErrInvalidRequestBody is returned when the request body is invalid
	ErrInvalidRequestBody = errors.New("invalid request body")

	// ErrTokenNotFound is returned when the token is not found
	ErrTokenNotFound = errors.New("token not found")

	// ErrMissingAuthHeader is returned when the auth header is missing
	ErrMissingAuthHeader = errors.New("missing authorization header")

	// ErrInvalidToken is returned when the token is invalid or expired
	ErrInvalidToken = errors.New("invalid or expired token")
)
