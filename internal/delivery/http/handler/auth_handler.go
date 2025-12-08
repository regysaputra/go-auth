package handler

import (
	"auth/internal/delivery/http/helper"
	"auth/internal/delivery/http/response"
	"auth/internal/usecase"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Web string const
const Web = "web"

// AuthHandler represents the auth handler object
type AuthHandler struct {
	logger                         *slog.Logger
	loginUserUseCase               *usecase.LoginUserUseCase
	refreshTokenUseCase            *usecase.RefreshTokenUseCase
	verifyEmailUseCase             *usecase.VerifyTokenUseCase
	requestPasswordResetUseCase    *usecase.RequestPasswordResetUseCase
	resetPasswordUseCase           *usecase.ResetPasswordUseCase
	requestVerificationCodeUseCase *usecase.RequestVerificationCodeUseCase
	verifyCodeUseCase              *usecase.VerifyCodeUseCase
	requestLoginOTPUseCase         *usecase.RequestLoginOTPUseCase
	verifyLoginOTPUseCase          *usecase.VerifyLoginOTPUseCase
	logoutUseCase                  *usecase.LogoutUseCase
}

// NewAuthHandler creates a new auth handler object
func NewAuthHandler(
	logger *slog.Logger,
	loginUC *usecase.LoginUserUseCase,
	refreshUC *usecase.RefreshTokenUseCase,
	verifyUC *usecase.VerifyTokenUseCase,
	requestPasswordResetUC *usecase.RequestPasswordResetUseCase,
	resetPasswordUC *usecase.ResetPasswordUseCase,
	requestVerificationCodeUC *usecase.RequestVerificationCodeUseCase,
	verifyCodeUC *usecase.VerifyCodeUseCase,
	requestLoginOTPUC *usecase.RequestLoginOTPUseCase,
	verifyLoginOTPUC *usecase.VerifyLoginOTPUseCase,
) *AuthHandler {
	return &AuthHandler{
		logger:                         logger,
		loginUserUseCase:               loginUC,
		refreshTokenUseCase:            refreshUC,
		verifyEmailUseCase:             verifyUC,
		requestPasswordResetUseCase:    requestPasswordResetUC,
		resetPasswordUseCase:           resetPasswordUC,
		requestVerificationCodeUseCase: requestVerificationCodeUC,
		verifyCodeUseCase:              verifyCodeUC,
		requestLoginOTPUseCase:         requestLoginOTPUC,
		verifyLoginOTPUseCase:          verifyLoginOTPUC,
	}
}

// LoginRequest represent the request body for a login user
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

// PasswordResetRequest represent the request body for password reset
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represent the request body for reset password
type ResetPasswordRequest struct {
	Password string `json:"password"`
}

// RequestVerificationCodeRequest represent the request body for request code
type RequestVerificationCodeRequest struct {
	Email string `json:"email"`
}

// VerifyCodeRequest represent the request body for verify code
type VerifyCodeRequest struct {
	Code string `json:"code"`
}

// RequestLoginOTPRequest represent the request body for request login otp
type RequestLoginOTPRequest struct {
	Email string `json:"email"`
}

// LogoutRequest represent the request body for request logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LoginUser represent the login user handler
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"), h.logger)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ipAddress := helper.ExtractIP(r)
	result, err := h.loginUserUseCase.Execute(r.Context(), req.Email, req.Password, req.RememberMe, userAgent, ipAddress)

	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else if errors.Is(err, usecase.ErrInvalidCredentials) {
			response.WriteError(w, http.StatusUnauthorized, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}
		return
	}

	// Check client type
	clientType := r.Header.Get("X-Client-Type")

	if clientType == Web || clientType == "" {
		setRefreshCookie(w, result.RefreshToken, result.RefreshTokenExp)
		response.WriteSuccess(w, http.StatusOK, map[string]any{
			"user":         result.LoginUserObj,
			"access_token": result.AccessToken,
		})

		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"user":              result.LoginUserObj,
		"access_token":      result.AccessToken,
		"refresh_token":     result.RefreshToken,
		"refresh_token_exp": result.RefreshTokenExp,
	})
}

// RefreshToken represent the refresh token handler
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie (web)
	var refreshToken string

	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	}

	// Get token for a non-web client
	if refreshToken == "" {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return
		}

		refreshToken = req.RefreshToken
	}

	// write an error response if a refresh token is not found in both clients
	if refreshToken == "" {
		response.WriteError(w, http.StatusUnauthorized, usecase.ErrTokenNotFound, h.logger)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ipAddress := helper.ExtractIP(r)

	// Call the use case to perform the refresh logic
	result, err := h.refreshTokenUseCase.Execute(r.Context(), refreshToken, userAgent, ipAddress)

	if err != nil {
		clearRefreshCookie(w)
		response.WriteError(w, http.StatusUnauthorized, usecase.ErrInvalidToken, h.logger)
		return
	}

	// Check client type
	clientType := r.Header.Get("X-Client-Type")

	if clientType == Web || clientType == "" {
		setRefreshCookie(w, result.RefreshToken, result.RefreshTokenExp)
		response.WriteSuccess(w, http.StatusOK, map[string]any{
			"access_token": result.AccessToken,
		})

		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"access_token":      result.AccessToken,
		"refresh_token":     result.RefreshToken,
		"refresh_token_exp": result.RefreshTokenExp,
	})
}

// VerifyEmail represent the verify email handler
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get token from the query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrTokenNotFound, h.logger)
		return
	}

	// Get user agent and ip address
	userAgent := r.Header.Get("User-Agent")
	ipAddress := helper.ExtractIP(r)

	// Call the use case
	result, err := h.verifyEmailUseCase.Execute(r.Context(), token, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			response.WriteError(w, http.StatusUnauthorized, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	// Check client type
	clientType := r.Header.Get("X-Client-Type")

	if clientType == Web || clientType == "" {
		setRefreshCookie(w, result.RefreshToken, result.RefreshTokenExp)
		response.WriteSuccess(w, http.StatusOK, map[string]any{
			"user":         result.LoginUserObj,
			"access_token": result.AccessToken,
		})

		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"user":              result.LoginUserObj,
		"access_token":      result.AccessToken,
		"refresh_token":     result.RefreshToken,
		"refresh_token_exp": result.RefreshTokenExp,
	})
}

// RequestPasswordReset represent the request password reset handler
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrInvalidRequestBody, h.logger)
		return
	}

	err := h.requestPasswordResetUseCase.Execute(r.Context(), req.Email)
	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	// Always return a positive-like response to prevent email enumeration attack
	response.WriteSuccess(w, http.StatusAccepted, map[string]string{"message": "a password reset link has been sent"})
}

// ResetPassword represent the reset password handler
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrTokenNotFound, h.logger)
		return
	}

	// Get user password
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrInvalidRequestBody, h.logger)
		return
	}

	// Execute a reset password use case
	err := h.resetPasswordUseCase.Execute(r.Context(), token, req.Password)
	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else if errors.Is(err, usecase.ErrInvalidToken) {
			response.WriteError(w, http.StatusUnauthorized, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "a password has been reset successfully"})
}

// RequestVerificationCode represent the request verification code handler
func (h *AuthHandler) RequestVerificationCode(w http.ResponseWriter, r *http.Request) {
	// Get request body
	var req RequestVerificationCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrInvalidRequestBody, h.logger)
		return
	}

	// Execute use case
	err := h.requestVerificationCodeUseCase.Execute(r.Context(), req.Email)
	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else if errors.Is(err, usecase.ErrEmailExists) {
			response.WriteError(w, http.StatusConflict, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	response.WriteSuccess(w, http.StatusAccepted, map[string]string{"message": "a verification code has been sent to your email"})
}

// VerifyCode represent the verify code handler
func (h *AuthHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	// Get code from a request body
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrInvalidRequestBody, h.logger)
		return
	}

	// Execute use case
	token, err := h.verifyCodeUseCase.Execute(r.Context(), req.Code)
	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else if errors.Is(err, usecase.ErrInvalidVerificationCode) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"registration_token": token,
	})
}

// RequestLoginOTP handles the request to send a one-time password (OTP) to a user's email for login purposes.
func (h *AuthHandler) RequestLoginOTP(w http.ResponseWriter, r *http.Request) {
	var req RequestLoginOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrInvalidRequestBody, h.logger)
		return
	}

	err := h.requestLoginOTPUseCase.Execute(r.Context(), req.Email)
	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	response.WriteSuccess(w, http.StatusAccepted, map[string]string{"message": "a login otp has been sent to your email"})
}

// VerifyLoginOTP the handler for verifying a login otp and authenticating the user
func (h *AuthHandler) VerifyLoginOTP(w http.ResponseWriter, r *http.Request) {
	// Get code from a request body
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrInvalidRequestBody, h.logger)
		return
	}

	// Get user agent and ip address
	userAgent := r.Header.Get("User-Agent")
	ipAddress := helper.ExtractIP(r)

	// Execute use case
	result, err := h.verifyLoginOTPUseCase.Execute(r.Context(), req.Code, userAgent, ipAddress)
	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else if errors.Is(err, usecase.ErrInvalidVerificationCode) {
			response.WriteError(w, http.StatusUnprocessableEntity, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	// Check client type
	clientType := r.Header.Get("X-Client-Type")

	if clientType == Web || clientType == "" {
		setRefreshCookie(w, result.RefreshToken, result.RefreshTokenExp)
		response.WriteSuccess(w, http.StatusOK, map[string]any{
			"user":         result.LoginUserObj,
			"access_token": result.AccessToken,
		})

		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"user":              result.LoginUserObj,
		"access_token":      result.AccessToken,
		"refresh_token":     result.RefreshToken,
		"refresh_token_exp": result.RefreshTokenExp,
	})
}

// Logout handles user logout by invalidating the provided refresh token and clearing the refresh token cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, usecase.ErrInvalidRequestBody, h.logger)
		return
	}

	err := h.logoutUseCase.Execute(r.Context(), req.RefreshToken)
	if err != nil {
		var validationErrors *usecase.ValidationErrors

		if errors.As(err, &validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else if errors.Is(err, usecase.ErrInvalidToken) {
			response.WriteError(w, http.StatusUnauthorized, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	clearRefreshCookie(w)
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "you have been logged out successfully"})
}

// Helper function
func setRefreshCookie(w http.ResponseWriter, refreshToken string, expiresAt time.Time) {
	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}

func clearRefreshCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
