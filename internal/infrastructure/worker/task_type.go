package worker

// This file defines the names/types of all our background tasks
const (
	TypeSendEmailVerificationLink  = "email:verify_token"
	TypeSendEmailPasswordResetLink = "email:password_reset"
	TypeSendEmailVerificationCode  = "email:verify_code"
	TypeSendEmailLoginOTP          = "email:login_otp"
)
