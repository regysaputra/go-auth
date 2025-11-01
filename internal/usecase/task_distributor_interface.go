package usecase

import (
	"context"
)

type TaskDistributor interface {
	DistributeTaskSendEmailVerificationLink(ctx context.Context, email string, token string) error
	DistributeTaskSendEmailPasswordResetLink(ctx context.Context, email string, token string) error
	DistributeTaskSendEmailVerificationCode(ctx context.Context, email string, code string) error
	DistributeTaskSendEmailLoginOTP(ctx context.Context, email string, code string) error
}
