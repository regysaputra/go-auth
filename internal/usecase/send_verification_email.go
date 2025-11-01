package usecase

import (
	"context"
	"time"
)

type SendEmailVerificationLinkUseCase struct {
	verifyRepository VerificationTokenRepository
	taskDistributor  TaskDistributor
}

func NewSendEmailVerificationLinkUseCase(
	verifyRepository VerificationTokenRepository,
	taskDistributor TaskDistributor,
) *SendEmailVerificationLinkUseCase {
	return &SendEmailVerificationLinkUseCase{
		verifyRepository: verifyRepository,
		taskDistributor:  taskDistributor,
	}
}

func (uc *SendEmailVerificationLinkUseCase) Execute(ctx context.Context, userID int64, email string) error {
	// Generate verification token
	rawToken, err := uc.verifyRepository.Generate()
	if err != nil {
		return err
	}

	// Hash the token
	tokenHash := uc.verifyRepository.Hash(rawToken)

	// Save the hash token to repository
	err = uc.verifyRepository.Save(ctx, userID, tokenHash, time.Hour)
	if err != nil {
		return err
	}

	err = uc.taskDistributor.DistributeTaskSendEmailVerificationLink(ctx, email, rawToken)

	if err != nil {
		return err
	}

	return nil
}
