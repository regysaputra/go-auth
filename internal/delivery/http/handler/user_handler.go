package handler

import (
	"auth/internal/delivery/http/helper"
	"auth/internal/delivery/http/middleware"
	"auth/internal/delivery/http/response"
	"auth/internal/usecase"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// UserHandler represents the user handler object
type UserHandler struct {
	logger                      *slog.Logger
	registerUserUseCase         *usecase.RegisterUserUseCase
	registerUserWithCodeUseCase *usecase.RegisterUserWithCodeUseCase
	getUserProfileUseCase       *usecase.GetUserProfileUseCase
}

// NewUserHandler creates a new user handler object
func NewUserHandler(
	logger *slog.Logger,
	registerUC *usecase.RegisterUserUseCase,
	registerWithCodeUC *usecase.RegisterUserWithCodeUseCase,
	getProfileUC *usecase.GetUserProfileUseCase,
) *UserHandler {
	return &UserHandler{
		logger:                      logger,
		registerUserUseCase:         registerUC,
		registerUserWithCodeUseCase: registerWithCodeUC,
		getUserProfileUseCase:       getProfileUC,
	}
}

// RegisterUserRequest represent the request body for register user
type RegisterUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterUserWithCodeRequest represent the request body for register user
type RegisterUserWithCodeRequest struct {
	Name              string `json:"name"`
	Password          string `json:"password"`
	VerificationToken string `json:"verification_token"`
}

// RegisterUser represent the register user handler
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"), h.logger)
		return
	}

	// Call register use case
	err := h.registerUserUseCase.Execute(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		// Determine status code based on an error type
		var validationErrors *usecase.ValidationErrors

		if errors.Is(err, validationErrors) {
			response.WriteError(w, http.StatusBadRequest, err, h.logger)
		} else if errors.Is(err, usecase.ErrEmailExists) {
			response.WriteError(w, http.StatusConflict, err, h.logger)
		} else {
			response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		}

		return
	}

	response.WriteSuccess(w, http.StatusCreated, map[string]string{
		"message": "user created successfully, please check your email to verify your account",
	})
}

// RegisterUserWithCode represent the register user with code handler
func (h *UserHandler) RegisterUserWithCode(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserWithCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("invalid request body"), h.logger)
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ipAddress := helper.ExtractIP(r)

	// Register user
	result, err := h.registerUserWithCodeUseCase.Execute(r.Context(), req.Name, req.Password, req.VerificationToken, userAgent, ipAddress)

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

	// Check client type
	clientType := r.Header.Get("X-Client-Type")

	if clientType == "web" || clientType == "" {
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

// GetUserProfile represent the get user profile handler
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, err, h.logger)
		return
	}

	// Call use case
	result, err := h.getUserProfileUseCase.Execute(ctx, userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err, h.logger)
		return
	}

	// Send a successful response
	response.WriteSuccess(w, http.StatusOK, result)
}
