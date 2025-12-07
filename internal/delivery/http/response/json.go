package response

import (
	"auth/internal/usecase"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// --- Standardized JSend Response Structs ---

// SuccessResponse follows the JSend spec for successful requests.
type SuccessResponse struct {
	Status string `json:"status" example:"success"`
	Data   any    `json:"data"`
}

// JSendFailResponse follows the JSend spec for client-side errors (4xx).
// The 'data' field contains the validation errors.
type JSendFailResponse struct {
	Status string `json:"status" example:"fail"`
	Data   any    `json:"data"`
}

// JSendErrorResponse follows the JSend spec for server-side errors (5xx).
// The 'message' field contains a description of the error.
type JSendErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"an unexpected error occurred"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return
	}
}

// WriteSuccess is a smart helper that formats the response according to JSend specs.
func WriteSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, SuccessResponse{
		Status: "success",
		Data:   data,
	})
}

// WriteError is a smart helper that formats the response according to JSend specs
// based on the type of error it receives.
func WriteError(w http.ResponseWriter, status int, err error, logger *slog.Logger) {
	var validationErrs *usecase.ValidationErrors

	// Check if this is a structured validation error
	if errors.As(err, &validationErrs) {
		// It's "fail" (4xx) with structured data
		writeJSON(w, status, JSendFailResponse{
			Status: "fail",
			Data:   validationErrs.Fields,
		})
		return
	}

	// Check if it's a simple, known business logic failure
	if status >= 400 && status < 500 {
		// It's "fail" (4xx) with a simple message
		writeJSON(w, status, JSendFailResponse{
			Status: "fail",
			Data:   map[string]string{"message": err.Error()},
		})

		return
	}

	// Otherwise, it's an unexpected "error" (5xx)
	// Log the full error for debugging
	logger.Error("internal server error", "error", err)
	writeJSON(w, status, JSendErrorResponse{
		Status:  "error",
		Message: "internal server error",
	})
}
