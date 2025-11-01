package handler

import (
	"auth/internal/usecase"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type AuthHandler struct {
	logger                         *slog.Logger
	loginUserUseCase               *usecase.LoginUserUseCase
	refreshTokenUseCase            *usecase.RefreshTokenUseCase
	verifyEmailUseCase             *usecase.VerifyEmailUseCase
	requestPasswordResetUseCase    *usecase.RequestPasswordResetUseCase
	resetPasswordUseCase           *usecase.ResetPasswordUseCase
	requestVerificationCodeUseCase *usecase.RequestVerificationCodeUseCase
	verifyCodeUseCase              *usecase.VerifyCodeUseCase
	requestLoginOTPUseCase         *usecase.RequestLoginOTPUseCase
	verifyLoginOTPUseCase          *usecase.VerifyLoginOTPUseCase
}

func NewAuthHandler(
	logger *slog.Logger,
	loginUC *usecase.LoginUserUseCase,
	refreshUC *usecase.RefreshTokenUseCase,
	verifyUC *usecase.VerifyEmailUseCase,
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

type LoginUserRequest struct {
	Email      string `json:"email" example:"username@domain"`
	Password   string `json:"password" example:"password"`
	RememberMe bool   `json:"remember_me" example:"true"`
}

type LoginUserSuccessResponse struct {
	AccessToken   string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs"`
	RememberToken string `json:"remember_token" example:"InR5cCI6IkpXVCJ9eyJhbGciOiJIUzI1NiIs"`
}

type LoginUserFailResponse struct {
	Email    []string `json:"email" example:"email is required,email is invalid"`
	Password []string `json:"password" example:"password is required,password too short"`
}

type RefreshTokenResponse struct {
	AccessToken   string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs"`
	RememberToken string `json:"remember_token" example:"InR5cCI6IkpXVCJ9eyJhbGciOiJIUzI1NiIs"`
}

type PasswordResetRequest struct {
	Email string `json:"email" example:"username@domain"`
}

type PasswordResetSuccessResponse struct {
	Message string `json:"message" example:"Password reset link has been sent to your email"`
}

type PasswordResetFailResponse struct {
	Email []string `json:"email" example:"email is required,email is invalid"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" example:"$fesf&idsie94"`
}

type ResetPasswordSuccessResponse struct {
	Message string `json:"message" example:"Password has been reset successfully"`
}

type ResetPasswordFailResponse struct {
	Password []string `json:"password" example:"password is required,password too short"`
}

type RequestCodeRequest struct {
	Email string `json:"email" example:"username@domain"`
}

type RequestCodeSuccessResponse struct {
	Message string `json:"message" example:"A verification code has been to your email"`
}

type RequestCodeFailResponse struct {
	Message []string `json:"message" example:"email is required,email is invalid"`
}

type VerifyCodeRequest struct {
	Code string `json:"code" example:"123456"`
}

type VerifyCodeSuccessResponse struct {
	VerificationToken string `json:"verification_token" example:"eyHUhjgtIG"`
}

type VerifyCodeFailResponse struct {
	Code []string `json:"code" example:"code is required"`
}

type RequestLoginOTPRequest struct {
	Email string `json:"email" example:"username@domain"`
}

type RequestLoginOTPSuccessResponse struct {
	Message string `json:"message" example:"A verification code has been sent to your email"`
}

// LoginUser godoc
// @Summary			Logs in a user
// @Description  Authenticates a user by email and password and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body LoginUserRequest true "User Login Credentials"
// @Success      200 {object} SuccessResponse{data=LoginUserSuccessResponse}
// @Failure      400 {object} FailResponse{data=LoginUserFailResponse}
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/auth [post]
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	result, err := h.loginUserUseCase.Execute(r.Context(), req.Email, req.Password, req.RememberMe)

	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, usecase.ErrInvalidCredentials.Error())
		} else {
			writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
		}

		return
	}

	response := LoginUserSuccessResponse{AccessToken: result.AccessToken}

	if result.RememberToken != "" {
		setRememberCookie(w, result.RememberToken)    // for web client
		response.RememberToken = result.RememberToken // for non-web client
	}

	writeSuccess(w, http.StatusOK, response)
}

// RefreshToken godoc
// @Summary		Refreshes a user's session
// @Description Uses a remember token to generate a new JWT and a new remember token. The token can be provided via a cookie (for web) or an X-Remember-Token header (for non-web client)
// @Tags		auth
// @produce		json
// @Param        X-Remember-Token header string false "Remember Me Token for non-web clients"
// @Success      200 {object} SuccessResponse{data=RefreshTokenResponse}
// @Failure      401 {object} ErrorResponse
// @Failure      500 {object} ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie (web)
	rawToken := ""
	cookie, err := r.Cookie("remember_token")
	if err == nil {
		rawToken = cookie.Value
	}

	// Get token from header (non-web)
	if rawToken == "" {
		rawToken = r.Header.Get("X-Remember-Token")
	}

	// write error response if remember token is not found in both client
	if rawToken == "" {
		writeError(w, http.StatusUnauthorized, ErrTokenNotFound.Error())
		return
	}

	// Call the use case to perform the refresh logic
	result, err := h.refreshTokenUseCase.Execute(r.Context(), rawToken)
	if err != nil {
		clearRememberCookie(w)
		writeError(w, http.StatusUnauthorized, usecase.ErrInvalidToken.Error())
		return
	}

	// For web client, set the new remember token in a new cookie
	setRememberCookie(w, result.NewRememberToken)

	// For all clients: Send the new JWT and new remember token in the response body
	response := RefreshTokenResponse{
		AccessToken:   result.NewJWT,
		RememberToken: result.NewRememberToken,
	}

	writeSuccess(w, http.StatusOK, response)
}

// VerifyEmail godoc
// @Summary		Verifies a user's email
// @Description Uses a verification token from a query parameter to verify a user's email and log them in
// @Tags		auth
// @Produce		json
// @Param		token query string true "Email verification token"
// @Success     200 {object} SuccessResponse{data=LoginUserSuccessResponse}
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/auth/verify [get]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get token from the query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, ErrTokenNotFound.Error())
		return
	}

	// Call the use case
	result, err := h.verifyEmailUseCase.Execute(r.Context(), token)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, usecase.ErrInvalidCredentials.Error())
		} else {
			writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
		}
		return
	}

	// Set the remember me cookie and return the JWT
	setRememberCookie(w, result.RememberToken)
	response := LoginUserSuccessResponse{
		AccessToken:   result.AccessToken,
		RememberToken: result.RememberToken,
	}

	writeSuccess(w, http.StatusOK, response)
}

// RequestPasswordReset godoc
// @Summary		Request a password reset
// @Description Send a password reset link to the user email address
// @Tags		auth
// @Produce		json
// @Param		email body PasswordResetRequest true "User email"
// @Success 202 {object} SuccessResponse{data=PasswordResetSuccessResponse}
// @Failure 400 {object} FailResponse{data=PasswordResetFailResponse}
// @Failure 500 {object} ErrorResponse
// @Router	/api/v1/auth/password/request-reset [post]
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
	}

	err := h.requestPasswordResetUseCase.Execute(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
	}

	// Always return a positive-like response to prevent email enumeration attack
	writeSuccess(w, http.StatusAccepted, map[string]string{"message": "a password reset link has been sent"})
}

// ResetPassword godoc
// @Summary		Reset user password
// @Description Set a new password for the user using a valid reset token
// @Tags		auth
// @Produce		json
// @Param		password body ResetPasswordRequest true "User password"
// @Success 200 {object} SuccessResponse{data=ResetPasswordSuccessResponse}
// @Failure 400 {object} FailResponse{data=ResetPasswordFailResponse}
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router	/api/v1/auth/password/reset [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, ErrTokenNotFound.Error())
		return
	}

	// Get user password
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	// Execute reset password use case
	err := h.resetPasswordUseCase.Execute(r.Context(), token, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidToken) {
			writeError(w, http.StatusUnauthorized, usecase.ErrInvalidToken.Error())
			return
		}

		if errors.Is(err, usecase.ErrPasswordTooShort) {
			writeFail(w, http.StatusBadRequest, map[string][]string{"password": {usecase.ErrPasswordTooShort.Error()}})
			return
		}

		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
	}

	writeSuccess(w, http.StatusOK, map[string]string{"message": "a password has been reset successfully"})
}

// RequestVerificationCode godoc
// @Summary		Request a verification code
// @Description Send a 6-digit verification code to email
// @Tags		auth
// @Produce		json
// @Param		email body RequestCodeRequest true "Email to verify"
// @Success 202 {object} SuccessResponse{data=RequestCodeSuccessResponse}
// @Failure 400 {object} FailResponse{data=RequestCodeFailResponse}
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router	/api/v1/auth/request-code [post]
func (h *AuthHandler) RequestVerificationCode(w http.ResponseWriter, r *http.Request) {
	// Get request body
	var req RequestCodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	// Execute use case
	err := h.requestVerificationCodeUseCase.Execute(r.Context(), req.Email)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		if errors.Is(err, usecase.ErrEmptyEmail) {
			writeFail(w, http.StatusBadRequest, map[string]interface{}{
				"email": usecase.ErrEmptyEmail.Error(),
			})
			return
		}

		if errors.Is(err, usecase.ErrInvalidEmail) {
			writeError(w, http.StatusBadRequest, usecase.ErrInvalidEmail.Error())
			return
		}

		if errors.Is(err, usecase.ErrEmailExists) {
			writeError(w, http.StatusConflict, usecase.ErrEmailExists.Error())
			return
		}

		h.logger.Error("Failed to send verification code : ", "error", err)

		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
		return
	}

	response := RequestCodeSuccessResponse{
		Message: "a verification code has been sent to your email",
	}

	writeSuccess(w, http.StatusAccepted, response)
}

// VerifyCode godoc
// @Summary		Verify code
// @Description Verify code from email
// @Tags		auth
// @Produce		json
// @Param		code body VerifyCodeRequest true "Code to verify"
// @Success 200 {object} SuccessResponse{data=VerifyCodeSuccessResponse}
// @Failure 400 {object} FailResponse{data=VerifyCodeFailResponse}
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router	/api/v1/auth/verify-code [post]
func (h *AuthHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	// Get code from request body
	var req VerifyCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	// Execute use case
	token, err := h.verifyCodeUseCase.Execute(r.Context(), req.Code)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidVerificationCode) {
			writeError(w, http.StatusUnprocessableEntity, usecase.ErrInvalidVerificationCode.Error())
			return
		}

		h.logger.Error("Failed to verify code : ", "error", err)
		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
		return
	}

	response := VerifyCodeSuccessResponse{
		VerificationToken: token,
	}

	writeSuccess(w, http.StatusOK, response)
}

func (h *AuthHandler) RequestLoginOTP(w http.ResponseWriter, r *http.Request) {
	var req RequestLoginOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestBody.Error())
		return
	}

	err := h.requestLoginOTPUseCase.Execute(r.Context(), req.Email)
	if err != nil {
		validationErrors := make(map[string][]string)
		if errors.Is(err, usecase.ErrEmptyEmail) {
			validationErrors["email"] = append(validationErrors["email"], usecase.ErrEmptyEmail.Error())
		}

		if errors.Is(err, usecase.ErrInvalidEmail) {
			validationErrors["email"] = append(validationErrors["email"], usecase.ErrInvalidEmail.Error())
		}

		if len(validationErrors) > 0 {
			writeFail(w, http.StatusBadRequest, validationErrors)
			return
		}

		h.logger.Error("Failed to send OTP login : ", "error", err)
		writeError(w, http.StatusInternalServerError, usecase.ErrInternalServer.Error())
	}
}

// Helper function
func setRememberCookie(w http.ResponseWriter, rememberToken string) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    rememberToken,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)

}

func clearRememberCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
