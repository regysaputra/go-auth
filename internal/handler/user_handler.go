package handler

import (
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
	Name     string `json:"name" example:"Egi"`
	Email    string `json:"email" example:"username@domain"`
	Password string `json:"password" example:"@fg8s64gf!"`
}

// RegisterUserFailResponse represent the response body for register user fail
type RegisterUserFailResponse struct {
	Name     []string `json:"name" example:"name is required,"`
	Email    []string `json:"email" example:"email is required,email is invalid"`
	Password []string `json:"password" example:"password is required,password too short"`
}

// RegisterUserWithCodeSuccessResponse represent the response body for register user with code success
type RegisterUserWithCodeSuccessResponse struct {
	AccessToken   string `json:"access_token"`
	RememberToken string `json:"remember_token,omitempty"`
}

// RegisterUserSuccessResponse represent the response body for register user success
type RegisterUserSuccessResponse struct {
	Message string `json:"message"`
}

// RegisterUser godoc
// @Summary Register new user
// @Description Add new user
// @Tags user
// @Accept json
// @Produce json
// @Param user body handler.RegisterUserRequest true "User registration details"
// @Success 201 {object} SuccessResponse{data=RegisterUserSuccessResponse}
// @Failure 400 {object} FailResponse{data=RegisterUserFailResponse}
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	// Call register use case
	_, err := h.registerUserUseCase.Execute(r.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		validationErrors := make(map[string][]string)

		if errors.Is(err, usecase.ErrEmptyName) {
			validationErrors["name"] = append(validationErrors["name"], usecase.ErrEmptyName.Error())
		}
		if errors.Is(err, usecase.ErrEmptyEmail) {
			validationErrors["email"] = append(validationErrors["email"], usecase.ErrEmptyEmail.Error())
		}
		if errors.Is(err, usecase.ErrEmptyPassword) {
			validationErrors["password"] = append(validationErrors["password"], usecase.ErrEmptyPassword.Error())
		}
		if errors.Is(err, usecase.ErrInvalidEmail) {
			validationErrors["email"] = append(validationErrors["email"], usecase.ErrInvalidEmail.Error())
		}
		if errors.Is(err, usecase.ErrPasswordTooShort) {
			validationErrors["password"] = append(validationErrors["password"], usecase.ErrPasswordTooShort.Error())
		}
		if len(validationErrors) > 0 {
			writeFail(w, http.StatusBadRequest, validationErrors)
			return
		}

		if errors.Is(err, usecase.ErrEmailExists) {
			writeError(w, http.StatusConflict, usecase.ErrEmailExists.Error())
			return
		}

		h.logger.Error("Failed to register user : ", "error", err)

		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
		return
	}

	response := RegisterUserSuccessResponse{
		Message: "user created successfully, please check your email to verify your account",
	}

	// On success, return a 201 Created response.
	// The user's password field is automatically omitted by the `json:"-"` tag in the domain model.
	writeSuccess(w, http.StatusCreated, response)
}

// RegisterUserWithCode godoc
// @Summary Register new user
// @Description Add new user
// @Tags user
// @Accept json
// @Produce json
// @Param user body handler.RegisterUserRequest true "User registration details"
// @Success 201 {object} SuccessResponse{data=RegisterUserSuccessResponse}
// @Failure 400 {object} FailResponse{data=RegisterUserFailResponse}
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) RegisterUserWithCode(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	// Register user
	token, err := h.registerUserWithCodeUseCase.Execute(r.Context(), req.Name, req.Email, req.Password)

	if err != nil {
		validationErrors := make(map[string][]string)

		if errors.Is(err, usecase.ErrEmptyName) {
			validationErrors["name"] = append(validationErrors["name"], usecase.ErrEmptyName.Error())
		}
		if errors.Is(err, usecase.ErrEmptyEmail) {
			validationErrors["email"] = append(validationErrors["email"], usecase.ErrEmptyEmail.Error())
		}
		if errors.Is(err, usecase.ErrEmptyPassword) {
			validationErrors["password"] = append(validationErrors["password"], usecase.ErrEmptyPassword.Error())
		}
		if errors.Is(err, usecase.ErrInvalidEmail) {
			validationErrors["email"] = append(validationErrors["email"], usecase.ErrInvalidEmail.Error())
		}
		if errors.Is(err, usecase.ErrPasswordTooShort) {
			validationErrors["password"] = append(validationErrors["password"], usecase.ErrPasswordTooShort.Error())
		}
		if len(validationErrors) > 0 {
			writeFail(w, http.StatusBadRequest, validationErrors)
			return
		}

		if errors.Is(err, usecase.ErrEmailExists) {
			writeError(w, http.StatusConflict, usecase.ErrEmailExists.Error())
			return
		}

		h.logger.Error("Failed to register user : ", "error", err)

		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
		return
	}

	response := RegisterUserWithCodeSuccessResponse{
		AccessToken:   token.AccessToken,
		RememberToken: token.RememberToken,
	}

	// On success, return a 201 Created response.
	// The user's password field is automatically omitted by the `json:"-"` tag in the domain model.
	writeSuccess(w, http.StatusCreated, response)
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get detail user
// @Tags user
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} SuccessResponse{data=domain.User}
// @Failure 401 {object} ErrorResponse "unauthorized"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get userID
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, usecase.ErrUserUnauthorized.Error())
		return
	}

	// Call use case
	user, err := h.getUserProfileUseCase.Execute(ctx, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, usecase.ErrUserNotFound.Error())
			return
		}

		h.logger.Error("Failed to get user profile : ", "error", err)
		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
		return
	}

	// Send a successful response
	writeSuccess(w, http.StatusOK, user)
}
