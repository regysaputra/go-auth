package main

import (
	"auth/internal/infrastructure/service"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	SecretKey   string
	BaseURL     string
	SMTP        service.SMTPConfig

	// Token expiration time
	AccessTokenTTL        time.Duration
	RefreshTokenTTL       time.Duration
	PasswordResetTokenTTL time.Duration
	VerificationTokenTTL  time.Duration
	LoginOTPTTL           time.Duration
	VerificationCodeTTL   time.Duration
	RegistrationTokenTTL  time.Duration

	// Threshold
	RiskRevokeThreshold     int
	RiskInvalidateThreshold int
}

func loadConfig() (*Config, error) {
	// Load .env file if exists (for local development)
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}

		log.Println("Loaded .env file for local development")
	}

	config := &Config{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		SecretKey:   os.Getenv("SECRET_KEY"),
		BaseURL:     os.Getenv("BASE_URL"),
		SMTP: service.SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     os.Getenv("SMTP_PORT"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
			BaseURL:  os.Getenv("BASE_URL"),
		},
		AccessTokenTTL:          accessTokenExpiration,
		RefreshTokenTTL:         refreshTokenExpiration,
		PasswordResetTokenTTL:   passwordResetTokenExpiration,
		VerificationTokenTTL:    verificationTokenExpiration,
		LoginOTPTTL:             loginOTPExpiration,
		VerificationCodeTTL:     verificationCodeExpiration,
		RegistrationTokenTTL:    registrationTokenExpiration,
		RiskRevokeThreshold:     riskRevokeThreshold,
		RiskInvalidateThreshold: riskInvalidateThreshold,
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL environment variable not set")
	}

	if c.RedisURL == "" {
		return errors.New("REDIS_URL environment variable not set")
	}

	if c.SecretKey == "" {
		return errors.New("SECRET_KEY environment variable not set")
	}

	if c.BaseURL == "" {
		return errors.New("BASE_URL environment variable not set")
	}

	if c.Port == "" {
		return errors.New("PORT environment variable not set")
	}

	if c.SMTP.From == "" || c.SMTP.Port == "" || c.SMTP.Host == "" || c.SMTP.Password == "" || c.SMTP.Username == "" {
		return errors.New("SMTP configuration is not set")
	}

	return nil
}
