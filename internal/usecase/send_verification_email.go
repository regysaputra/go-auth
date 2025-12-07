package usecase

import (
	"context"
)

// SendEmailVerificationLinkUseCase represents the use case for sending an email verification link
type SendEmailVerificationLinkUseCase struct {
	tokenRepository  TokenRepository
	verifyRepository VerificationTokenRepository
	taskDistributor  TaskDistributor
}

// NewSendEmailVerificationLinkUseCase creates a new SendEmailVerificationLinkUseCase object
func NewSendEmailVerificationLinkUseCase(
	verifyRepository VerificationTokenRepository,
	taskDistributor TaskDistributor,
) *SendEmailVerificationLinkUseCase {
	return &SendEmailVerificationLinkUseCase{
		verifyRepository: verifyRepository,
		taskDistributor:  taskDistributor,
	}
}

// Execute executes the use case
func (uc *SendEmailVerificationLinkUseCase) Execute(ctx context.Context, userID int64, email string) error {
	// Generate verification token
	token, err := uc.tokenRepository.GenerateOpaqueToken()
	if err != nil {
		return err
	}

	// Hash the token
	tokenHash := uc.tokenRepository.HashToken(token)

	// Save the hash token to a repository
	err = uc.verifyRepository.Save(ctx, userID, tokenHash)
	if err != nil {
		return err
	}

	err = uc.taskDistributor.DistributeTaskSendEmailVerificationLink(ctx, email, token)

	if err != nil {
		return err
	}

	return nil
}
