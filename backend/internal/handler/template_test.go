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
	"github.com/sergiojoz/gym-tracker/internal/middlewares"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTemplateUseCase is a mock implementation of template use case for testing.
type MockTemplateUseCase struct {
	CreateFunc  func(ctx context.Context, t *domain.WorkoutTemplate) error
	UpdateFunc  func(ctx context.Context, t *domain.WorkoutTemplate) error
	DeleteFunc  func(ctx context.Context, userID, templateID uuid.UUID) error
	GetByIDFunc func(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error)
	ListFunc    func(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error)
}

func (m *MockTemplateUseCase) Create(ctx context.Context, t *domain.WorkoutTemplate) error {
	return m.CreateFunc(ctx, t)
}
func (m *MockTemplateUseCase) Update(ctx context.Context, t *domain.WorkoutTemplate) error {
	return m.UpdateFunc(ctx, t)
}
func (m *MockTemplateUseCase) Delete(ctx context.Context, userID, templateID uuid.UUID) error {
	return m.DeleteFunc(ctx, userID, templateID)
}
func (m *MockTemplateUseCase) GetByID(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error) {
	return m.GetByIDFunc(ctx, userID, templateID)
}
func (m *MockTemplateUseCase) List(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
	return m.ListFunc(ctx, userID, filter)
}

// contextWithUserID adds a user ID to the context (simulating auth middleware).
func contextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return middlewares.ContextWithUserID(ctx, userID)
}

func TestTemplateHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockUC := &MockTemplateUseCase{
		CreateFunc: func(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
			uid, ok := middlewares.UserIDFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, userID, uid)
			assert.Equal(t, "Push Day", tmpl.Name)
			tmpl.ID = uuid.New()
			return nil
		},
	}

	handler := NewTemplateHandler(mockUC)

	body := map[string]interface{}{
		"name":        "Push Day",
		"description": "Chest focus",
		"slots": []map[string]interface{}{
			{
				"exercise_id":  exerciseID.String(),
				"order":        0,
				"target_sets":  4,
				"target_reps":  10,
				"target_weight": 80.0,
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Push Day", resp["name"])
	assert.NotEmpty(t, resp["id"])
}

func TestTemplateHandler_Create_InvalidJSON(t *testing.T) {
	mockUC := &MockTemplateUseCase{}
	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTemplateHandler_Create_ValidationError(t *testing.T) {
	mockUC := &MockTemplateUseCase{
		CreateFunc: func(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
			return domain.NewValidationError("template validation failed", []domain.FieldError{
				{Field: "name", Message: "is required"},
			})
		},
	}

	handler := NewTemplateHandler(mockUC)

	body := map[string]interface{}{"name": ""}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTemplateHandler_List_Success(t *testing.T) {
	userID := uuid.New()
	templates := []*domain.WorkoutTemplate{
		{ID: uuid.New(), UserID: userID, Name: "T1"},
		{ID: uuid.New(), UserID: userID, Name: "T2"},
	}

	mockUC := &MockTemplateUseCase{
		ListFunc: func(ctx context.Context, uid uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, 10, filter.Limit)
			return templates, false, nil
		},
	}

	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates?limit=10", nil)
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

func TestTemplateHandler_List_Empty(t *testing.T) {
	mockUC := &MockTemplateUseCase{
		ListFunc: func(ctx context.Context, uid uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
			return []*domain.WorkoutTemplate{}, false, nil
		},
	}

	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	items, ok := resp["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 0)
}

func TestTemplateHandler_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()
	tmpl := &domain.WorkoutTemplate{
		ID:     templateID,
		UserID: userID,
		Name:   "My Template",
		Slots: []*domain.TemplateSlot{
			{ID: uuid.New(), ExerciseID: uuid.New(), Order: 0, TargetSets: 3},
		},
	}

	mockUC := &MockTemplateUseCase{
		GetByIDFunc: func(ctx context.Context, uid, tid uuid.UUID) (*domain.WorkoutTemplate, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, templateID, tid)
			return tmpl, nil
		},
	}

	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID.String(), nil)
	req = setURLParam(req, "id", templateID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, templateID.String(), resp["id"])
	assert.Equal(t, "My Template", resp["name"])
	slots, ok := resp["slots"].([]interface{})
	require.True(t, ok)
	assert.Len(t, slots, 1)
}

func TestTemplateHandler_GetByID_InvalidUUID(t *testing.T) {
	mockUC := &MockTemplateUseCase{}
	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/not-a-uuid", nil)
	req = setURLParam(req, "id", "not-a-uuid")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTemplateHandler_GetByID_NotFound(t *testing.T) {
	mockUC := &MockTemplateUseCase{
		GetByIDFunc: func(ctx context.Context, uid, tid uuid.UUID) (*domain.WorkoutTemplate, error) {
			return nil, domain.ErrNotFound
		},
	}

	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+uuid.New().String(), nil)
	req = setURLParam(req, "id", uuid.New().String())
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTemplateHandler_Update_Success(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()

	mockUC := &MockTemplateUseCase{
		UpdateFunc: func(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
			assert.Equal(t, templateID, tmpl.ID)
			assert.Equal(t, userID, tmpl.UserID)
			return nil
		},
	}

	handler := NewTemplateHandler(mockUC)

	body := map[string]interface{}{
		"name": "Updated",
		"slots": []map[string]interface{}{
			{"exercise_id": uuid.New().String(), "order": 0, "target_sets": 5, "target_reps": 5},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = setURLParam(req, "id", templateID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTemplateHandler_Update_NotFound(t *testing.T) {
	mockUC := &MockTemplateUseCase{
		UpdateFunc: func(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
			return domain.ErrNotFound
		},
	}

	handler := NewTemplateHandler(mockUC)

	body := map[string]interface{}{"name": "Ghost"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+uuid.New().String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = setURLParam(req, "id", uuid.New().String())
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTemplateHandler_Delete_Success(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()

	mockUC := &MockTemplateUseCase{
		DeleteFunc: func(ctx context.Context, uid, tid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, templateID, tid)
			return nil
		},
	}

	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+templateID.String(), nil)
	req = setURLParam(req, "id", templateID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTemplateHandler_Delete_NotFound(t *testing.T) {
	mockUC := &MockTemplateUseCase{
		DeleteFunc: func(ctx context.Context, uid, tid uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+uuid.New().String(), nil)
	req = setURLParam(req, "id", uuid.New().String())
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTemplateHandler_Delete_InvalidUUID(t *testing.T) {
	mockUC := &MockTemplateUseCase{}
	handler := NewTemplateHandler(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/not-a-uuid", nil)
	req = setURLParam(req, "id", "not-a-uuid")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
