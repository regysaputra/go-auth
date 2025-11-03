package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// SuccessResponse represents the success response
type SuccessResponse struct {
	Status string      `json:"status" example:"success"`
	Data   interface{} `json:"data"`
}

// FailResponse represents the fail response
type FailResponse struct {
	Status string      `json:"status" example:"fail"`
	Data   interface{} `json:"data"`
}

// ErrorResponse represents the error response
type ErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message"`
}

func writeSuccess(w http.ResponseWriter, httpStatus int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	response := SuccessResponse{Status: "success", Data: payload}
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error(
			"Failed to encode response",
			slog.Any("error", err),                    // Log the error object
			slog.Int("status", httpStatus),            // Log context
			slog.String("response_status", "success"), // Log context
		)

		return
	}
}

func writeFail(w http.ResponseWriter, httpStatus int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	response := FailResponse{Status: "fail", Data: payload}
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error(
			"Failed to encode response",
			slog.Any("error", err),                    // Log the error object
			slog.Int("status", httpStatus),            // Log context
			slog.String("response_status", "success"), // Log context
		)

		return
	}
}

func writeError(w http.ResponseWriter, httpStatus int, message string) {
	w.Header().Set("Content-Type", "application/json")
	response := ErrorResponse{Status: "error", Message: message}
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error(
			"Failed to encode response",
			slog.Any("error", err),                    // Log the error object
			slog.Int("status", httpStatus),            // Log context
			slog.String("response_status", "success"), // Log context
		)

		return
	}
}
