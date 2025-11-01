package usecase

type EmailSender interface {
	SendEmailVerificationLink(email string, token string) error
	SendEmailPasswordResetLink(email string, token string) error
	SendEmailVerificationCode(email string, code string) error
	SendEmailLoginOTP(email string, token string) error
}
