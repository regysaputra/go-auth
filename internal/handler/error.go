package handler

import "errors"

var (
	ErrMalformedAuthHeader = errors.New("malformed auth header")        // ErrMalformedAuthHeader is returned when the auth header is malformed
	ErrMethodNotAllowed    = errors.New("method not allowed")           // ErrMethodNotAllowed is returned when the method is not allowed
	ErrInvalidRequestBody  = errors.New("invalid request body")         // ErrInvalidRequestBody is returned when the request body is invalid
	ErrTokenNotFound       = errors.New("token not found")              // ErrTokenNotFound is returned when the token is not found
	ErrMissingAuthHeader   = errors.New("missing authorization header") // ErrMissingAuthHeader is returned when the auth header is missing
	ErrInvalidToken        = errors.New("invalid or expired token")     // ErrInvalidToken is returned when the token is invalid or expired
)
