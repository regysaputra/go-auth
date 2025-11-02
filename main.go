package main

import (
	_ "auth/docs"
	"auth/internal/handler"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/usecase"
	"auth/internal/worker"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Logger Adapter for Asynq

// SlogAsynqLogger is an adapter to make slog.Logger compatible with asynq.Logger
type SlogAsynqLogger struct {
	logger *slog.Logger
}

func NewSlogAsynqLogger(logger *slog.Logger) *SlogAsynqLogger {
	return &SlogAsynqLogger{logger: logger}
}

// Debug These methods implement the asynq.Logger interface
func (l *SlogAsynqLogger) Debug(args ...interface{}) {
	l.logger.Debug(fmt.Sprint(args...))
}

func (l *SlogAsynqLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *SlogAsynqLogger) Warn(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *SlogAsynqLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *SlogAsynqLogger) Fatal(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...)) // slog doesn't have Fatal, so we use Error
	os.Exit(1)
}

// @title			Authentication API
// @version			1.0
// @description		This is server for authentication system API
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @contact.url		http://www.egiwebdev.id/support
// @contact.email	egiwebdev@gmail.com
// @license.name	Apache 2.0
// @license.url		http://www.apache.org/licenses/LICENSE-2.0.html
// @host			localhost:8080
// @BasePath		/
// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Connect to the database
	dbpool, err := connectToDb()
	if err != nil {
		logger.Error("Could not connect to the database", "error", err)
		os.Exit(1)
	}

	defer dbpool.Close()
	logger.Info("Connected to database")

	// Initialize Asynq
	redisConnOpt, err := asynq.ParseRedisURI(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Error("Could not connect to redis url", "error", err)
		os.Exit(1)
	}

	asynqClient := asynq.NewClient(redisConnOpt)
	defer func(asynqClient *asynq.Client) {
		err := asynqClient.Close()
		if err != nil {
			logger.Error("Could not close asynq client", "error", err)
		}
	}(asynqClient)

	taskDistributor := worker.NewRedisTaskDistributor(asynqClient)

	asynqLogger := NewSlogAsynqLogger(logger)
	asynqServer := asynq.NewServer(redisConnOpt, asynq.Config{
		Logger: asynqLogger,
	})

	// Initialize service
	smtpConfig := service.SmtpConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		BaseURL:  os.Getenv("BASE_URL"),
	}
	emailSender := service.NewSmtpEmailSender(smtpConfig)

	// Initialize repositories
	userRepository := repository.NewPostgresUserRepository(dbpool)
	authRepository := repository.NewJWTAuthRepository()
	rememberRepository := repository.NewPostgresRememberTokenRepository(dbpool)
	verifyRepository := repository.NewPostgresVerificationTokenRepository(dbpool)
	passwordResetRepository := repository.NewPostgresPasswordResetTokenRepository(dbpool)
	emailVerificationCodeRepository := repository.NewPostgresEmailVerificationCodeRepository(dbpool)
	loginOTPRepository := repository.NewPostgresLoginOTPRepository(dbpool)

	// Initialize use case
	sendEmailVerificationLinkUseCase := usecase.NewSendEmailVerificationLinkUseCase(verifyRepository, taskDistributor)
	registerUserUseCase := usecase.NewRegisterUserUseCase(userRepository, sendEmailVerificationLinkUseCase)
	//sendVerificationEmail := usecase.NewSendEmailVerificationLinkUseCase(verifyRepository, taskDistributor)
	loginUseCase := usecase.NewLoginUserUseCase(userRepository, authRepository, rememberRepository)
	refreshTokenUseCase := usecase.NewRefreshTokenUseCase(userRepository, rememberRepository, authRepository)
	verifyEmailUseCase := usecase.NewVerifyEmailUseCase(userRepository, verifyRepository, loginUseCase)
	requestPasswordResetUseCase := usecase.NewRequestPasswordResetUseCase(logger, userRepository, passwordResetRepository, taskDistributor)
	resetPasswordUseCase := usecase.NewResetPasswordUseCase(userRepository, passwordResetRepository)
	requestVerificationCodeUseCase := usecase.NewRequestVerificationCodeUseCase(emailVerificationCodeRepository, userRepository, taskDistributor)
	verifyCodeUseCase := usecase.NewVerifyCodeUseCase(emailVerificationCodeRepository, authRepository)
	getUserProfileUseCase := usecase.NewGetUserProfileUseCase(userRepository)
	requestLoginOTPUseCase := usecase.NewRequestLoginOTPUseCase(logger, loginOTPRepository, userRepository, taskDistributor)
	verifyLoginOTPUseCase := usecase.NewVerifyLoginOTPUseCase(loginOTPRepository, userRepository, loginUseCase)
	registerUserWithCodeUseCase := usecase.NewRegisterUserWithCodeUseCase(userRepository, verifyCodeUseCase, loginUseCase)

	// Initialize handler
	userHandler := handler.NewUserHandler(logger, registerUserUseCase, registerUserWithCodeUseCase, getUserProfileUseCase)
	authHandler := handler.NewAuthHandler(
		logger,
		loginUseCase,
		refreshTokenUseCase,
		verifyEmailUseCase,
		requestPasswordResetUseCase,
		resetPasswordUseCase,
		requestVerificationCodeUseCase,
		verifyCodeUseCase,
		requestLoginOTPUseCase,
		verifyLoginOTPUseCase,
	)
	authMiddleware := handler.AuthMiddleware

	// Start task processor
	taskProcessor := worker.NewRedisTaskProcessor(asynqServer, emailSender, logger)
	go func() {
		err := taskProcessor.Start()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start task processor", "error", err)
		}
	}()

	// Setup router and middleware
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("BASE_URL")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Backend server is healthy"))
		if err != nil {
			logger.Error("Backend server is error:", "error", err)
		}
	})

	// API v1 routes
	router.Route("/api/v1", func(api chi.Router) {
		// Swagger documentation
		api.Get("/swagger/*", httpSwagger.WrapHandler)

		// Auth routes
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/", authHandler.LoginUser)
			auth.Post("/refresh", authHandler.RefreshToken)
			auth.Get("/verify-email", authHandler.VerifyEmail)
			auth.Post("/request-code", authHandler.RequestVerificationCode)
			auth.Post("/verify-code", authHandler.VerifyCode)
			auth.Post("/password/request-reset", authHandler.RequestPasswordReset)
			auth.Post("/password/reset", authHandler.ResetPassword)
			auth.Post("/otp/request", authHandler.RequestLoginOTP)
		})

		// User routes
		api.Route("/users", func(user chi.Router) {
			user.Post("/", userHandler.RegisterUser)

			// Protected routes
			user.Group(func(user chi.Router) {
				user.Use(authMiddleware)
				user.Get("/me", userHandler.GetUserProfile)
			})
		})
	})

	// Set up the server
	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	go func() {
		logger.Info("Starting the server on port :" + os.Getenv("PORT"))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server", "error", err)
	}

	logger.Info("HTTP server is shutting down")

	taskProcessor.Shutdown()
	logger.Info("Task processor shut down")
}

func connectToDb() (*pgxpool.Pool, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return nil, errors.New("DATABASE_URL environment variable not set")
	}

	pool, err := pgxpool.New(context.Background(), connString)

	if err != nil {
		return nil, err
	}
	return pool, nil
}
