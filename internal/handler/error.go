package handler

import "errors"

var (
	ErrMalformedAuthHeader = errors.New("malformed auth header")
	ErrMethodNotAllowed    = errors.New("method not allowed")
	ErrInvalidRequestBody  = errors.New("invalid request body")
	ErrTokenNotFound       = errors.New("token not found")
	ErrMissingAuthHeader   = errors.New("missing authorization header")
	ErrInvalidToken        = errors.New("invalid or expired token")
)
