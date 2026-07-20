package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	jwtPkg "github.com/sergiojoz/gym-tracker/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	userID := uuid.New()
	jwtCfg := &jwtPkg.Config{
		AccessSecret: "test-secret",
		AccessExpiry: 15 * time.Minute,
	}

	token, err := jwtPkg.GenerateAccessToken(userID, jwtCfg)
	require.NoError(t, err)

	middleware := AuthMiddleware(jwtCfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID, ok := UserIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, userID, ctxUserID)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	jwtCfg := &jwtPkg.Config{
		AccessSecret: "test-secret",
		AccessExpiry: 15 * time.Minute,
	}

	middleware := AuthMiddleware(jwtCfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtCfg := &jwtPkg.Config{
		AccessSecret: "test-secret",
		AccessExpiry: 15 * time.Minute,
	}

	middleware := AuthMiddleware(jwtCfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	jwtCfg := &jwtPkg.Config{
		AccessSecret: "test-secret",
		AccessExpiry: -1 * time.Minute, // Already expired
	}

	token, err := jwtPkg.GenerateAccessToken(userID, jwtCfg)
	require.NoError(t, err)

	middleware := AuthMiddleware(jwtCfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_WrongTokenType(t *testing.T) {
	userID := uuid.New()
	jwtCfg := &jwtPkg.Config{
		RefreshSecret: "test-secret",
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	// Generate a refresh token instead of access token
	token, err := jwtPkg.GenerateRefreshToken(userID, jwtCfg)
	require.NoError(t, err)

	// But use access secret for validation
	accessCfg := &jwtPkg.Config{
		AccessSecret: "test-secret",
		AccessExpiry: 15 * time.Minute,
	}

	middleware := AuthMiddleware(accessCfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MalformedAuthHeader(t *testing.T) {
	jwtCfg := &jwtPkg.Config{
		AccessSecret: "test-secret",
		AccessExpiry: 15 * time.Minute,
	}

	middleware := AuthMiddleware(jwtCfg)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "token"},
		{"empty after bearer", "Bearer "},
		{"wrong prefix", "Basic token"},
		{"multiple spaces", "Bearer  token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestUserIDFromContext_NotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := UserIDFromContext(req.Context())
	assert.False(t, ok)
}

func TestUserIDFromContext_Set(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), userIDKey, userID)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	ctxUserID, ok := UserIDFromContext(req.Context())
	assert.True(t, ok)
	assert.Equal(t, userID, ctxUserID)
}

func TestExtractToken_Valid(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"valid", "Bearer token123", "token123"},
		{"with spaces", "Bearer  token123", "token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := extractToken(tt.header)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestExtractToken_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"empty", ""},
		{"no bearer", "token123"},
		{"wrong prefix", "Basic token123"},
		{"empty after bearer", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := extractToken(tt.header)
			assert.Empty(t, token)
		})
	}
}

func TestExtractToken_CaseInsensitive(t *testing.T) {
	tests := []struct {
		header   string
		expected string
	}{
		{"bearer token123", "token123"},
		{"BEARER token123", "token123"},
		{"Bearer token123", "token123"},
	}

	for _, tt := range tests {
		token := extractToken(tt.header)
		assert.Equal(t, tt.expected, token)
	}
}

func TestExtractToken_WithWhitespace(t *testing.T) {
	token := extractToken("Bearer   token123  ")
	// Should trim and extract correctly
	assert.True(t, strings.Contains(token, "token123"))
}
