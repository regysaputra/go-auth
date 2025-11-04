package usecase

import "context"

// EmailSender interface
type EmailSender interface {
	SendEmailVerificationLink(ctx context.Context, email string, token string) error
	SendEmailPasswordResetLink(ctx context.Context, email string, token string) error
	SendEmailVerificationCode(ctx context.Context, email string, code string) error
	SendEmailLoginOTP(ctx context.Context, email string, token string) error
}
