package handler

import (
	"context"
	"net/http"

	"github.com/sergiojoz/gym-tracker/internal/usecase"
)

// AuthUseCaseInterface defines the interface for auth operations.
type AuthUseCaseInterface interface {
	Register(ctx context.Context, req *usecase.RegisterRequest) (*usecase.RegisterResponse, error)
	Login(ctx context.Context, req *usecase.LoginRequest) (*usecase.LoginResponse, error)
	Refresh(ctx context.Context, req *usecase.RefreshRequest) (*usecase.RefreshResponse, error)
	Logout(ctx context.Context, req *usecase.LogoutRequest) error
}

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authUC AuthUseCaseInterface
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authUC AuthUseCaseInterface) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// RegisterRequest represents the JSON request body for registration.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	ucReq := &usecase.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authUC.Register(r.Context(), ucReq)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp.User)
}

// LoginRequest represents the JSON request body for login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the JSON response body for login.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	ucReq := &usecase.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	resp, err := h.authUC.Login(r.Context(), ucReq)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

// RefreshRequest represents the JSON request body for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents the JSON response body for token refresh.
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	ucReq := &usecase.RefreshRequest{
		RefreshToken: req.RefreshToken,
	}

	resp, err := h.authUC.Refresh(r.Context(), ucReq)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, RefreshResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	})
}

// LogoutRequest represents the JSON request body for logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	ucReq := &usecase.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}

	if err := h.authUC.Logout(r.Context(), ucReq); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
