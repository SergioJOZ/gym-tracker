package domain

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "something went wrong",
		Status:  500,
	}

	if err.Error() != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", err.Error())
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		wantCode   string
		wantStatus int
	}{
		{"ErrNotFound", ErrNotFound, "NOT_FOUND", 404},
		{"ErrConflict", ErrConflict, "CONFLICT", 409},
		{"ErrUnauthorized", ErrUnauthorized, "UNAUTHORIZED", 401},
		{"ErrForbidden", ErrForbidden, "FORBIDDEN", 403},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("Code: got %q, want %q", tt.err.Code, tt.wantCode)
			}
			if tt.err.Status != tt.wantStatus {
				t.Errorf("Status: got %d, want %d", tt.err.Status, tt.wantStatus)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	details := []FieldError{
		{Field: "email", Message: "invalid format"},
		{Field: "password", Message: "too short"},
	}

	err := NewValidationError("invalid input", details)

	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("Code: got %q, want %q", err.Code, "VALIDATION_ERROR")
	}
	if err.Status != 400 {
		t.Errorf("Status: got %d, want %d", err.Status, 400)
	}
	if len(err.Details) != 2 {
		t.Fatalf("Details: got %d, want 2", len(err.Details))
	}
	if err.Details[0].Field != "email" {
		t.Errorf("Details[0].Field: got %q, want %q", err.Details[0].Field, "email")
	}
}

func TestAppError_Is(t *testing.T) {
	// Test that errors.Is works with sentinel errors
	wrapped := errors.Join(ErrNotFound, errors.New("user not found"))
	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("expected wrapped error to match ErrNotFound")
	}
}

func TestNewAppError(t *testing.T) {
	err := NewAppError("CUSTOM_ERROR", "custom message", 422)
	if err.Code != "CUSTOM_ERROR" {
		t.Errorf("Code: got %q, want %q", err.Code, "CUSTOM_ERROR")
	}
	if err.Message != "custom message" {
		t.Errorf("Message: got %q, want %q", err.Message, "custom message")
	}
	if err.Status != 422 {
		t.Errorf("Status: got %d, want %d", err.Status, 422)
	}
}
