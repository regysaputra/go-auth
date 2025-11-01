package worker

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
)

// RedisTaskDistributor is the concrete implementation using Redis
type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(client *asynq.Client) *RedisTaskDistributor {
	return &RedisTaskDistributor{client: client}
}

func (d *RedisTaskDistributor) DistributeTaskSendEmailVerificationLink(ctx context.Context, email string, token string) error {
	task, err := NewSendEmailVerificationLinkPayload(email, token)
	if err != nil {
		return err
	}

	// Process the task with medium priority, and retry up to 3 times
	_, err = d.client.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Timeout(1*time.Minute))

	return err
}

func (d *RedisTaskDistributor) DistributeTaskSendEmailPasswordResetLink(ctx context.Context, email string, token string) error {
	task, err := NewSendEmailPasswordResetLinkPayload(email, token)

	if err != nil {
		return err
	}

	_, err = d.client.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Timeout(1*time.Minute))

	return err
}

func (d *RedisTaskDistributor) DistributeTaskSendEmailVerificationCode(ctx context.Context, email string, code string) error {
	task, err := NewSendEmailVerificationCodePayload(email, code)

	if err != nil {
		return err
	}

	_, err = d.client.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Timeout(1*time.Minute))

	return err
}

func (d *RedisTaskDistributor) DistributeTaskSendEmailLoginOTP(ctx context.Context, email string, code string) error {
	task, err := NewSendEmailLoginOTPPayload(email, code)
	if err != nil {
		return err
	}

	_, err = d.client.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Timeout(1*time.Minute))
	return err
}
