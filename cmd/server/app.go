package main

import (
	"auth/internal/infrastructure/worker"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	config        *Config
	logger        *slog.Logger
	dbPool        *pgxpool.Pool
	server        *http.Server
	taskProcessor *worker.RedisTaskProcessor
	asynqClient   *asynq.Client
}

// NewApp creates and initializes a new application instance
func NewApp(config *Config, logger *slog.Logger) (*App, error) {
	app := &App{
		config: config,
		logger: logger,
	}

	// Initialize database
	dbPool, err := app.initDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	app.dbPool = dbPool
	logger.Info("Connected to database")

	// Initialize Asynq (Redis)
	asynqServer, asynqClient, taskDistributor, err := app.initAsynq()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Asynq: %w", err)
	}
	app.asynqClient = asynqClient
	logger.Info("Connected to Redis")

	// Initialize service
	services := app.initService(config, logger)

	// Initialize repository
	repositories := app.initRepositories()

	// Initialize use case
	useCase := app.initUseCase(repositories, taskDistributor, services.geoIPService)

	// Initialize handler
	handlers := app.initHandler(useCase, services.geoIPService)

	// Setup router
	router := app.setupRouter(handlers)

	// Create an HTTP server
	app.server = &http.Server{
		Addr:              ":" + config.Port,
		Handler:           router,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
		ReadHeaderTimeout: defaultHeaderTimeout,
	}

	// Initialize and start task processor
	taskProcessor := worker.NewRedisTaskProcessor(asynqServer, services.emailService, logger)
	app.taskProcessor = taskProcessor

	return app, nil
}

func (a *App) Start() error {
	// Start task processor in background
	go func() {
		a.logger.Info("Starting task processor...")
		if err := a.taskProcessor.Start(); err != nil {
			a.logger.Error("Failed to start task processor", "error", err)
		}
	}()

	// Start HTTP server
	go func() {
		a.logger.Info("Starting HTTP server...", "port", a.config.Port)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("Failed to start HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	return nil
}

// Shutdown waits for interrupt signal and performs graceful shutdown
func (a *App) Shutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Failed to shutdown HTTP server", "error", err)
	} else {
		a.logger.Info("HTTP server shutdown successfully")
	}

	a.taskProcessor.Shutdown()
	a.logger.Info("Task processor shutdown successfully")
}

// Cleanup closes all open connections and resources
func (a *App) Cleanup() {
	if a.dbPool != nil {
		a.dbPool.Close()
		a.logger.Info("Database connection closed")
	}

	if a.asynqClient != nil {
		if err := a.asynqClient.Close(); err != nil {
			a.logger.Error("Failed to close asynq client", "error", err)
		} else {
			a.logger.Info("Asynq client closed")
		}
	}
}
