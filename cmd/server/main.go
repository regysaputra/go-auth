// Package main is the entry point of the application.
package main

import (
	"log/slog"
	"os"
	"time"
)

const (
	// Server settings
	defaultReadTimeout   = 5 * time.Second
	defaultWriteTimeout  = 10 * time.Second
	defaultIdleTimeout   = 120 * time.Second
	defaultHeaderTimeout = 3 * time.Second
	shutdownTimeout      = 10 * time.Second

	// Token expiration time
	accessTokenExpiration        = 5 * time.Minute
	refreshTokenExpiration       = 24 * time.Hour
	passwordResetTokenExpiration = time.Hour
	verificationTokenExpiration  = time.Hour
	loginOTPExpiration           = 10 * time.Minute
	verificationCodeExpiration   = 10 * time.Minute
	registrationTokenExpiration  = time.Hour

	// File paths
	geoIPCityPath = "./pkg/geoip/GeoLite2-City.mmdb"
	geoIPASNPath  = "./pkg/geoip/GeoLite2-ASN.mmdb"
	openAPIPath   = "./public/openapi.yaml"

	// Thresholds
	riskRevokeThreshold     = 70
	riskInvalidateThreshold = 40
)

func main() {
	// Initialize logger
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config, err := loadConfig()
	if err != nil {
		slogger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	slogger.Info("Loaded config", "config", config)

	// Initialize application
	app, err := NewApp(config, slogger)
	if err != nil {
		slogger.Error("Failed to initialize app", "error", err)
		os.Exit(1)
	}
	defer app.Cleanup()

	// Start server
	if err := app.Start(); err != nil {
		slogger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	app.Shutdown()
}
