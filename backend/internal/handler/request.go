package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sergiojoz/gym-tracker/internal/domain"
)

// DecodeJSONBody decodes a JSON request body into the provided struct.
func DecodeJSONBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return domain.NewAppError("INVALID_REQUEST", "request body is empty", http.StatusBadRequest)
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return domain.NewAppError("INVALID_JSON", "invalid JSON format", http.StatusBadRequest)
	}

	return nil
}

// ParsePaginationParams extracts cursor and limit from query parameters.
func ParsePaginationParams(r *http.Request) (cursor string, limit int) {
	cursor = r.URL.Query().Get("cursor")
	
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limit = 20 // default
	} else {
		// Simple parsing, no error handling for invalid values
		limit = 0
		for _, c := range limitStr {
			if c >= '0' && c <= '9' {
				limit = limit*10 + int(c-'0')
			} else {
				limit = 20 // default on invalid
				break
			}
		}
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
	}

	return cursor, limit
}
