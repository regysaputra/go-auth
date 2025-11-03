package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// SendEmailVerificationLinkPayload is the data needed for the TypeSendEmailVerificationLink task
type SendEmailVerificationLinkPayload struct {
	Email string
	Token string
}

// NewSendEmailVerificationLinkPayload creates a new SendEmailVerificationLinkPayload object
func NewSendEmailVerificationLinkPayload(email string, token string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendEmailVerificationLinkPayload{
		Email: email,
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeSendEmailVerificationLink, payload), nil
}

// SendEmailPasswordResetLinkPayload is the data needed for the TypeSendEmailPasswordResetLink task
type SendEmailPasswordResetLinkPayload struct {
	Email string
	Token string
}

// NewSendEmailPasswordResetLinkPayload creates a new SendEmailPasswordResetLinkPayload object
func NewSendEmailPasswordResetLinkPayload(email string, token string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendEmailPasswordResetLinkPayload{
		Email: email,
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeSendEmailPasswordResetLink, payload), nil
}

// SendEmailVerificationCodePayload is the data needed for the TypeSendEmailVerificationCode task
type SendEmailVerificationCodePayload struct {
	Email string
	Code  string
}

// NewSendEmailVerificationCodePayload creates a new SendEmailVerificationCodePayload object
func NewSendEmailVerificationCodePayload(email string, code string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendEmailVerificationCodePayload{
		Email: email,
		Code:  code,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeSendEmailVerificationCode, payload), nil
}

// SendLoginOTPPayload is the data needed for the TypeSendEmailLoginOTP task
type SendLoginOTPPayload struct {
	Email string
	Code  string
}

// NewSendEmailLoginOTPPayload creates a new SendLoginOTPPayload object
func NewSendEmailLoginOTPPayload(email string, code string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendLoginOTPPayload{
		Email: email,
		Code:  code,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeSendEmailLoginOTP, payload), nil
}
