package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuthUseCase is a mock implementation of AuthUseCase for testing.
type MockAuthUseCase struct {
	RegisterFunc func(ctx context.Context, req *usecase.RegisterRequest) (*usecase.RegisterResponse, error)
	LoginFunc    func(ctx context.Context, req *usecase.LoginRequest) (*usecase.LoginResponse, error)
	RefreshFunc  func(ctx context.Context, req *usecase.RefreshRequest) (*usecase.RefreshResponse, error)
	LogoutFunc   func(ctx context.Context, req *usecase.LogoutRequest) error
}

func (m *MockAuthUseCase) Register(ctx context.Context, req *usecase.RegisterRequest) (*usecase.RegisterResponse, error) {
	return m.RegisterFunc(ctx, req)
}

func (m *MockAuthUseCase) Login(ctx context.Context, req *usecase.LoginRequest) (*usecase.LoginResponse, error) {
	return m.LoginFunc(ctx, req)
}

func (m *MockAuthUseCase) Refresh(ctx context.Context, req *usecase.RefreshRequest) (*usecase.RefreshResponse, error) {
	return m.RefreshFunc(ctx, req)
}

func (m *MockAuthUseCase) Logout(ctx context.Context, req *usecase.LogoutRequest) error {
	return m.LogoutFunc(ctx, req)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	userID := uuid.New()
	mockUC := &MockAuthUseCase{
		RegisterFunc: func(ctx context.Context, req *usecase.RegisterRequest) (*usecase.RegisterResponse, error) {
			return &usecase.RegisterResponse{
				User: &domain.User{
					ID:    userID,
					Email: req.Email,
				},
			}, nil
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", resp["email"])
}

func TestAuthHandler_Register_InvalidInput(t *testing.T) {
	mockUC := &MockAuthUseCase{
		RegisterFunc: func(ctx context.Context, req *usecase.RegisterRequest) (*usecase.RegisterResponse, error) {
			return nil, domain.NewValidationError("invalid input", []domain.FieldError{
				{Field: "email", Message: "invalid email"},
			})
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"email":    "invalid",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	userID := uuid.New()
	mockUC := &MockAuthUseCase{
		LoginFunc: func(ctx context.Context, req *usecase.LoginRequest) (*usecase.LoginResponse, error) {
			return &usecase.LoginResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				User: &domain.User{
					ID:    userID,
					Email: req.Email,
				},
			}, nil
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "access-token", resp["access_token"])
	assert.Equal(t, "refresh-token", resp["refresh_token"])
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockUC := &MockAuthUseCase{
		LoginFunc: func(ctx context.Context, req *usecase.LoginRequest) (*usecase.LoginResponse, error) {
			return nil, domain.ErrUnauthorized
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	mockUC := &MockAuthUseCase{
		RefreshFunc: func(ctx context.Context, req *usecase.RefreshRequest) (*usecase.RefreshResponse, error) {
			return &usecase.RefreshResponse{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
			}, nil
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"refresh_token": "old-refresh-token",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", resp["access_token"])
	assert.Equal(t, "new-refresh-token", resp["refresh_token"])
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	mockUC := &MockAuthUseCase{
		RefreshFunc: func(ctx context.Context, req *usecase.RefreshRequest) (*usecase.RefreshResponse, error) {
			return nil, domain.ErrUnauthorized
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"refresh_token": "invalid-token",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Refresh(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockUC := &MockAuthUseCase{
		LogoutFunc: func(ctx context.Context, req *usecase.LogoutRequest) error {
			return nil
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"refresh_token": "valid-token",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_Logout_InvalidToken(t *testing.T) {
	mockUC := &MockAuthUseCase{
		LogoutFunc: func(ctx context.Context, req *usecase.LogoutRequest) error {
			return domain.ErrUnauthorized
		},
	}

	handler := NewAuthHandler(mockUC)

	body := map[string]string{
		"refresh_token": "invalid-token",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
