package main

import (
	"auth/internal/delivery/http/handler"
	"auth/internal/infrastructure/logger"
	"auth/internal/infrastructure/repository"
	"auth/internal/infrastructure/service"
	"auth/internal/infrastructure/worker"
	"auth/internal/usecase"
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	user                  usecase.UserRepository
	token                 usecase.TokenRepository
	refreshToken          usecase.RefreshTokenRepository
	verificationToken     usecase.VerificationTokenRepository
	passwordReset         usecase.PasswordResetTokenRepository
	emailVerificationCode usecase.EmailVerificationCodeRepository
	loginOTP              usecase.LoginOTPRepository
}

type Service struct {
	emailService *service.SMTPEmailService
	geoIPService *service.GeoIPService
}

type UseCase struct {
	registerUser              *usecase.RegisterUserUseCase
	registerUserWithCode      *usecase.RegisterUserWithCodeUseCase
	login                     *usecase.LoginUserUseCase
	refreshToken              *usecase.RefreshTokenUseCase
	verifyEmail               *usecase.VerifyTokenUseCase
	requestPasswordReset      *usecase.RequestPasswordResetUseCase
	resetPassword             *usecase.ResetPasswordUseCase
	requestVerificationCode   *usecase.RequestVerificationCodeUseCase
	verifyCode                *usecase.VerifyCodeUseCase
	getUserProfile            *usecase.GetUserProfileUseCase
	requestLoginOTP           *usecase.RequestLoginOTPUseCase
	verifyLoginOTP            *usecase.VerifyLoginOTPUseCase
	sendEmailVerificationLink *usecase.SendEmailVerificationLinkUseCase
}

type Handler struct {
	user *handler.UserHandler
	auth *handler.AuthHandler
	test *handler.TestHandler
}

func (a *App) initDatabase() (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), a.config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func (a *App) initAsynq() (*asynq.Server, *asynq.Client, usecase.TaskDistributor, error) {
	// Initialize Asynq
	redisConnOpt, err := asynq.ParseRedisURI(a.config.RedisURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not connect to redis url: %w", err)
	}

	asynqClient := asynq.NewClient(redisConnOpt)
	taskDistributor := worker.NewRedisTaskDistributor(asynqClient)

	asynqLogger := logger.NewSlogAsynqLogger(a.logger)
	asynqServer := asynq.NewServer(redisConnOpt, asynq.Config{
		Logger: asynqLogger,
	})

	return asynqServer, asynqClient, taskDistributor, nil
}

func (a *App) initService(config *Config, logger *slog.Logger) *Service {
	return &Service{
		emailService: service.NewSMTPEmailService(config.SMTP, logger),
		geoIPService: service.MustNewGeoIPService(geoIPCityPath, geoIPASNPath),
	}
}

func (a *App) initRepositories() *Repository {
	return &Repository{
		user:                  repository.NewPostgresUserRepository(a.dbPool),
		token:                 repository.NewTokenRepository(a.config.SecretKey),
		refreshToken:          repository.NewPostgresRefreshTokenRepository(a.dbPool, a.config.RefreshTokenTTL),
		verificationToken:     repository.NewPostgresVerificationTokenRepository(a.dbPool, a.config.VerificationTokenTTL),
		passwordReset:         repository.NewPostgresPasswordResetTokenRepository(a.dbPool, a.config.PasswordResetTokenTTL),
		emailVerificationCode: repository.NewPostgresEmailVerificationCodeRepository(a.dbPool, a.config.VerificationCodeTTL),
		loginOTP:              repository.NewPostgresLoginOTPRepository(a.dbPool, a.config.LoginOTPTTL),
	}
}

func (a *App) initUseCase(repository *Repository, taskDistributor usecase.TaskDistributor, geoIPService *service.GeoIPService) *UseCase {
	sendEmailVerificationLink := usecase.NewSendEmailVerificationLinkUseCase(repository.verificationToken, taskDistributor)
	login := usecase.NewLoginUserUseCase(repository.user, repository.token, repository.refreshToken, geoIPService, a.config.AccessTokenTTL, a.config.RefreshTokenTTL)
	verifyCode := usecase.NewVerifyCodeUseCase(repository.emailVerificationCode, repository.token, a.config.RegistrationTokenTTL)

	return &UseCase{
		sendEmailVerificationLink: sendEmailVerificationLink,
		registerUser:              usecase.NewRegisterUserUseCase(repository.user, sendEmailVerificationLink),
		login:                     login,
		refreshToken:              usecase.NewRefreshTokenUseCase(repository.user, repository.refreshToken, repository.token, geoIPService, a.logger, a.config.AccessTokenTTL, a.config.RiskRevokeThreshold, a.config.RiskInvalidateThreshold),
		verifyEmail:               usecase.NewVerifyTokenUseCase(repository.user, repository.verificationToken, login),
		requestPasswordReset:      usecase.NewRequestPasswordResetUseCase(a.logger, repository.user, repository.passwordReset, repository.token, taskDistributor),
		resetPassword:             usecase.NewResetPasswordUseCase(repository.user, repository.token, repository.passwordReset),
		requestVerificationCode:   usecase.NewRequestVerificationCodeUseCase(repository.emailVerificationCode, repository.user, taskDistributor),
		verifyCode:                verifyCode,
		getUserProfile:            usecase.NewGetUserProfileUseCase(repository.user),
		requestLoginOTP:           usecase.NewRequestLoginOTPUseCase(a.logger, repository.loginOTP, repository.user, taskDistributor),
		verifyLoginOTP:            usecase.NewVerifyLoginOTPUseCase(repository.loginOTP, repository.user, login),
		registerUserWithCode:      usecase.NewRegisterUserWithCodeUseCase(repository.user, verifyCode, login),
	}
}

func (a *App) initHandler(useCase *UseCase, geoIPService *service.GeoIPService) *Handler {
	return &Handler{
		user: handler.NewUserHandler(
			a.logger,
			useCase.registerUser,
			useCase.registerUserWithCode,
			useCase.getUserProfile,
		),
		auth: handler.NewAuthHandler(
			a.logger,
			useCase.login,
			useCase.refreshToken,
			useCase.verifyEmail,
			useCase.requestPasswordReset,
			useCase.resetPassword,
			useCase.requestVerificationCode,
			useCase.verifyCode,
			useCase.requestLoginOTP,
			useCase.verifyLoginOTP,
		),
		test: handler.NewTestHandler(
			a.logger,
			geoIPService,
		),
	}
}
