package worker

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// This file define the data structure for our task payloads

// SendEmailVerificationLinkPayload is the data needed for the TypeSendEmailVerificationLink task
type SendEmailVerificationLinkPayload struct {
	Email string
	Token string
}

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

type SendEmailPasswordResetLinkPayload struct {
	Email string
	Token string
}

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

type SendEmailVerificationCodePayload struct {
	Email string
	Code  string
}

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

type SendLoginOTPPayload struct {
	Email string
	Code  string
}

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
