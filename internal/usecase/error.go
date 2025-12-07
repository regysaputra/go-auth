package usecase

import "errors"

// Pre-defined errors for specific business rule violations.
var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidToken            = errors.New("invalid or expired token")
	ErrEmailExists             = errors.New("user with this email already exists")
	ErrInvalidVerificationCode = errors.New("invalid or expired verification code")
	ErrInternalServer          = errors.New("internal server error")
	ErrUserNotFound            = errors.New("user not found")
	ErrSuspiciousActivity      = errors.New("suspicious activity detected")
	ErrTokenNotFound           = errors.New("token not found")
	ErrMalformedAuthHeader     = errors.New("malformed auth header")
	ErrInvalidRequestBody      = errors.New("invalid request body")
	ErrMissingAuthHeader       = errors.New("missing authorization header")
)

// ValidationErrors is custom error type to hold field-specific validation errors.
type ValidationErrors struct {
	Fields map[string][]string
}

// NewValidationErrors creates a new ValidationErrors instance.
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		Fields: make(map[string][]string),
	}
}

// Add adds a new validation error for a given field.
func (ve *ValidationErrors) Add(field, message string) {
	ve.Fields[field] = append(ve.Fields[field], message)
}

// HasErrors returns true if there are any validation errors.
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.Fields) > 0
}

// Error implements the error interface, allowing this struct to be returned as an error.
func (ve *ValidationErrors) Error() string {
	return "validation error"
}
