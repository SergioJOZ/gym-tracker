package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	jwtPkg "github.com/sergiojoz/gym-tracker/pkg/jwt"
)

func TestSetupRouter(t *testing.T) {
	// Create mock handlers
	mockAuthUC := &MockAuthUseCase{}
	mockExerciseUC := &MockExerciseUseCase{}
	mockTemplateUC := &MockTemplateUseCase{}
	mockSessionUC := &MockSessionUseCase{}
	mockProgressUC := &MockProgressUseCase{}

	authHandler := NewAuthHandler(mockAuthUC)
	exerciseHandler := NewExerciseHandler(mockExerciseUC)
	templateHandler := NewTemplateHandler(mockTemplateUC)
	sessionHandler := NewSessionHandler(mockSessionUC)
	progressHandler := NewProgressHandler(mockProgressUC)
	mediaHandler := NewMediaHandler("/tmp/media", "gifs", "thumbnails")

	jwtCfg := &jwtPkg.Config{
		AccessSecret:  "test-secret",
		RefreshSecret: "test-refresh-secret",
	}

	router := SetupRouter(authHandler, exerciseHandler, templateHandler, sessionHandler, progressHandler, mediaHandler, jwtCfg)

	// Test that router is not nil
	assert.NotNil(t, router)

	// Test health endpoint (should require auth)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test public auth endpoint
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Should not be 404 (route exists)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}
