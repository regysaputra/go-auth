package worker

import (
	"auth/internal/usecase"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
}

// RedisTaskProcessor is the concrete implementation for processing tasks from Redis
type RedisTaskProcessor struct {
	server      *asynq.Server
	emailSender usecase.EmailSender
	logger      *slog.Logger
}

func NewRedisTaskProcessor(server *asynq.Server, emailSender usecase.EmailSender, logger *slog.Logger) *RedisTaskProcessor {
	return &RedisTaskProcessor{
		server:      server,
		emailSender: emailSender,
		logger:      logger,
	}
}

// Start register task handler and start the Asynq server
func (p *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeSendEmailVerificationLink, p.handleTaskSendEmailVerificationLink)
	mux.HandleFunc(TypeSendEmailPasswordResetLink, p.handleTaskSendEmailPasswordResetLink)
	mux.HandleFunc(TypeSendEmailVerificationCode, p.handleTaskSendEmailVerificationCode)

	p.logger.Info("Starting task processor...")

	return p.server.Start(mux)
}

func (p *RedisTaskProcessor) Shutdown() {
	p.logger.Info("Shutting down task processor")
	p.server.Shutdown()
}

func (p *RedisTaskProcessor) handleTaskSendEmailVerificationLink(ctx context.Context, t *asynq.Task) error {
	var payload SendEmailVerificationLinkPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		p.logger.Error("Failed to unmarshal verification email payload", "error", err)
		return err
	}

	p.logger.Info("Processing verification email task", "email", payload.Email)
	return p.emailSender.SendEmailVerificationLink(payload.Email, payload.Token)
}

func (p *RedisTaskProcessor) handleTaskSendEmailPasswordResetLink(ctx context.Context, t *asynq.Task) error {
	var payload SendEmailPasswordResetLinkPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		p.logger.Error("Failed to unmarshal verification email payload", "error", err)
		return err
	}

	p.logger.Info("Processing password reset email task", "email", payload.Email)
	return p.emailSender.SendEmailPasswordResetLink(payload.Email, payload.Token)
}

func (p *RedisTaskProcessor) handleTaskSendEmailVerificationCode(ctx context.Context, t *asynq.Task) error {
	var payload SendEmailVerificationCodePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		p.logger.Error("Failed to unmarshal verification code payload", "error", err)
		return err
	}

	p.logger.Info("Processing verification code task", "email", payload.Email)
	return p.emailSender.SendEmailVerificationCode(payload.Email, payload.Code)
}
