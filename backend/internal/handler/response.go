package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sergiojoz/gym-tracker/internal/domain"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// ErrorResponse represents an error response body.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody represents the error details.
type ErrorBody struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Details []domain.FieldError `json:"details,omitempty"`
}

// respondError writes an error response based on the AppError type.
func respondError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError

	// Check if it's an AppError
	if e, ok := err.(*domain.AppError); ok {
		appErr = e
	} else {
		// Default to internal error
		appErr = &domain.AppError{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
			Status:  http.StatusInternalServerError,
		}
	}

	response := ErrorResponse{
		Error: ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}

	respondJSON(w, appErr.Status, response)
}
