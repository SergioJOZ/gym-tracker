package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/middlewares"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSessionUseCase is a mock implementation of session use case for testing.
type MockSessionUseCase struct {
	CreateFunc  func(ctx context.Context, s *domain.WorkoutSession) error
	UpdateFunc  func(ctx context.Context, s *domain.WorkoutSession) error
	DeleteFunc  func(ctx context.Context, userID, sessionID uuid.UUID) error
	GetByIDFunc func(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error)
	ListFunc    func(ctx context.Context, userID uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error)
}

func (m *MockSessionUseCase) Create(ctx context.Context, s *domain.WorkoutSession) error {
	return m.CreateFunc(ctx, s)
}
func (m *MockSessionUseCase) Update(ctx context.Context, s *domain.WorkoutSession) error {
	return m.UpdateFunc(ctx, s)
}
func (m *MockSessionUseCase) Delete(ctx context.Context, userID, sessionID uuid.UUID) error {
	return m.DeleteFunc(ctx, userID, sessionID)
}
func (m *MockSessionUseCase) GetByID(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error) {
	return m.GetByIDFunc(ctx, userID, sessionID)
}
func (m *MockSessionUseCase) List(ctx context.Context, userID uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error) {
	return m.ListFunc(ctx, userID, filter)
}

func TestSessionHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockUC := &MockSessionUseCase{
		CreateFunc: func(ctx context.Context, session *domain.WorkoutSession) error {
			uid, ok := middlewares.UserIDFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, userID, uid)
			assert.Equal(t, "Morning Push", session.Name)
			session.ID = uuid.New()
			return nil
		},
	}

	handler := NewSessionHandler(mockUC)

	body := map[string]interface{}{
		"name":     "Morning Push",
		"notes":    "Felt strong",
		"start_at": time.Now().Format(time.RFC3339),
		"exercises": []map[string]interface{}{
			{
				"exercise_id": exerciseID.String(),
				"order":       0,
				"notes":       "Bench press",
				"sets": []map[string]interface{}{
					{"order": 0, "weight": 80.0, "reps": 10},
					{"order": 1, "weight": 85.0, "reps": 8},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Morning Push", resp["name"])
	assert.NotEmpty(t, resp["id"])
}

func TestSessionHandler_Create_InvalidJSON(t *testing.T) {
	mockUC := &MockSessionUseCase{}
	handler := NewSessionHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_Create_ValidationError(t *testing.T) {
	mockUC := &MockSessionUseCase{
		CreateFunc: func(ctx context.Context, session *domain.WorkoutSession) error {
			return domain.NewValidationError("session validation failed", []domain.FieldError{
				{Field: "name", Message: "is required"},
			})
		},
	}

	handler := NewSessionHandler(mockUC)

	body := map[string]interface{}{"name": ""}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_List_Success(t *testing.T) {
	userID := uuid.New()
	sessions := []*domain.WorkoutSession{
		{ID: uuid.New(), UserID: userID, Name: "S1"},
		{ID: uuid.New(), UserID: userID, Name: "S2"},
	}

	mockUC := &MockSessionUseCase{
		ListFunc: func(ctx context.Context, uid uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, 10, filter.Limit)
			return sessions, false, nil
		},
	}

	handler := NewSessionHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions?limit=10", nil)
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	items, ok := resp["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 2)
	assert.Equal(t, false, resp["has_more"])
}

func TestSessionHandler_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	session := &domain.WorkoutSession{
		ID:     sessionID,
		UserID: userID,
		Name:   "My Session",
		Exercises: []*domain.SessionExercise{
			{ID: uuid.New(), ExerciseID: uuid.New(), Order: 0},
		},
	}

	mockUC := &MockSessionUseCase{
		GetByIDFunc: func(ctx context.Context, uid, sid uuid.UUID) (*domain.WorkoutSession, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, sessionID, sid)
			return session, nil
		},
	}

	handler := NewSessionHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/"+sessionID.String(), nil)
	req = setURLParam(req, "id", sessionID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, sessionID.String(), resp["id"])
	assert.Equal(t, "My Session", resp["name"])
}

func TestSessionHandler_GetByID_InvalidUUID(t *testing.T) {
	mockUC := &MockSessionUseCase{}
	handler := NewSessionHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sessions/not-a-uuid", nil)
	req = setURLParam(req, "id", "not-a-uuid")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionHandler_Update_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	exerciseID := uuid.New()

	mockUC := &MockSessionUseCase{
		UpdateFunc: func(ctx context.Context, session *domain.WorkoutSession) error {
			assert.Equal(t, sessionID, session.ID)
			assert.Equal(t, userID, session.UserID)
			return nil
		},
	}

	handler := NewSessionHandler(mockUC)

	body := map[string]interface{}{
		"name":     "Updated Session",
		"start_at": time.Now().Format(time.RFC3339),
		"exercises": []map[string]interface{}{
			{
				"exercise_id": exerciseID.String(),
				"order":       0,
				"sets": []map[string]interface{}{
					{"order": 0, "reps": 10},
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/sessions/"+sessionID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = setURLParam(req, "id", sessionID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSessionHandler_Delete_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	mockUC := &MockSessionUseCase{
		DeleteFunc: func(ctx context.Context, uid, sid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, sessionID, sid)
			return nil
		},
	}

	handler := NewSessionHandler(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/"+sessionID.String(), nil)
	req = setURLParam(req, "id", sessionID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSessionHandler_Delete_NotFound(t *testing.T) {
	mockUC := &MockSessionUseCase{
		DeleteFunc: func(ctx context.Context, uid, sid uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	handler := NewSessionHandler(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sessions/"+uuid.New().String(), nil)
	req = setURLParam(req, "id", uuid.New().String())
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
