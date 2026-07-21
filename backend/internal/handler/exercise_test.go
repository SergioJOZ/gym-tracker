package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExerciseUseCase is a mock implementation of exercise use case for testing.
type MockExerciseUseCase struct {
	ListFunc    func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error)
	GetByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Exercise, error)
}

func (m *MockExerciseUseCase) List(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
	return m.ListFunc(ctx, filter)
}

func (m *MockExerciseUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
	return m.GetByIDFunc(ctx, id)
}

func TestExerciseHandler_List_Success(t *testing.T) {
	exercises := []*domain.Exercise{
		{ID: uuid.New(), NameByLang: map[string]string{"en": "Bench Press"}, MuscleGroup: "chest", Difficulty: "intermediate"},
		{ID: uuid.New(), NameByLang: map[string]string{"en": "Squat"}, MuscleGroup: "legs", Difficulty: "intermediate"},
	}

	mockUC := &MockExerciseUseCase{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			assert.Equal(t, "chest", filter.MuscleGroup)
			return exercises, false, nil
		},
	}

	handler := NewExerciseHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/exercises?muscle_group=chest", nil)
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

func TestExerciseHandler_List_WithFilters(t *testing.T) {
	mockUC := &MockExerciseUseCase{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			assert.Equal(t, "bench", filter.Search)
			assert.Equal(t, "barbell", filter.Equipment)
			assert.Equal(t, "advanced", filter.Difficulty)
			assert.Equal(t, "strength", filter.Category)
			assert.Equal(t, 10, filter.Limit)
			return []*domain.Exercise{}, false, nil
		},
	}

	handler := NewExerciseHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/exercises?search=bench&equipment=barbell&difficulty=advanced&category=strength&limit=10", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExerciseHandler_List_WithCursor(t *testing.T) {
	mockUC := &MockExerciseUseCase{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			assert.Equal(t, "some-cursor", filter.Cursor)
			assert.Equal(t, 5, filter.Limit)
			return []*domain.Exercise{}, false, nil
		},
	}

	handler := NewExerciseHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/exercises?cursor=some-cursor&limit=5", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExerciseHandler_GetByID_Success(t *testing.T) {
	exerciseID := uuid.New()
	exercise := &domain.Exercise{
		ID:                 exerciseID,
		NameByLang:         map[string]string{"en": "Bench Press"},
		DescriptionsByLang: map[string]string{"en": "Compound chest exercise"},
		MuscleGroup:        "chest",
		Equipment:          "barbell",
		Difficulty:         "intermediate",
		Category:           "strength",
	}

	mockUC := &MockExerciseUseCase{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
			assert.Equal(t, exerciseID, id)
			return exercise, nil
		},
	}

	handler := NewExerciseHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/exercises/"+exerciseID.String(), nil)
	req = setURLParam(req, "id", exerciseID.String())
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, exerciseID.String(), resp["id"])
	// name_by_lang is a nested object {"en": "Bench Press"}, verify it exists
	nameByLang, ok := resp["name_by_lang"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Bench Press", nameByLang["en"])
}

func TestExerciseHandler_GetByID_NotFound(t *testing.T) {
	mockUC := &MockExerciseUseCase{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
			return nil, domain.ErrNotFound
		},
	}

	handler := NewExerciseHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/exercises/"+uuid.New().String(), nil)
	req = setURLParam(req, "id", uuid.New().String())
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestExerciseHandler_GetByID_InvalidUUID(t *testing.T) {
	mockUC := &MockExerciseUseCase{}
	handler := NewExerciseHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/exercises/not-a-uuid", nil)
	req = setURLParam(req, "id", "not-a-uuid")
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// setURLParam is a test helper to set chi URL parameters.
func setURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}
