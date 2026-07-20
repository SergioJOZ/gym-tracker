package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProgressUseCase is a mock implementation of progress use case for testing.
type MockProgressUseCase struct {
	ListPRsFunc          func(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error)
	ExerciseHistoryFunc  func(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error)
	SummaryFunc          func(ctx context.Context, userID uuid.UUID) (*repository.ProgressSummary, error)
}

func (m *MockProgressUseCase) ListPRs(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error) {
	return m.ListPRsFunc(ctx, userID)
}

func (m *MockProgressUseCase) ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error) {
	return m.ExerciseHistoryFunc(ctx, userID, exerciseID, cursor, limit)
}

func (m *MockProgressUseCase) Summary(ctx context.Context, userID uuid.UUID) (*repository.ProgressSummary, error) {
	return m.SummaryFunc(ctx, userID)
}

func TestProgressHandler_ListPRs_Success(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockUC := &MockProgressUseCase{
		ListPRsFunc: func(ctx context.Context, uid uuid.UUID) ([]*domain.PersonalRecord, error) {
			assert.Equal(t, userID, uid)
			return []*domain.PersonalRecord{
				{
					ID:         uuid.New(),
					UserID:     userID,
					ExerciseID: exerciseID,
					MaxWeight:  floatPtr(100.0),
					MaxReps:    intPtr(5),
				},
			}, nil
		},
	}

	handler := NewProgressHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/records", nil)
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.ListPRs(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []*domain.PersonalRecord
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, userID, resp[0].UserID)
}

func TestProgressHandler_ListPRs_Unauthorized(t *testing.T) {
	mockUC := &MockProgressUseCase{}
	handler := NewProgressHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/records", nil)
	w := httptest.NewRecorder()

	handler.ListPRs(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProgressHandler_ExerciseHistory_Success(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockUC := &MockProgressUseCase{
		ExerciseHistoryFunc: func(ctx context.Context, uid, eid uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, exerciseID, eid)
			assert.Equal(t, "", cursor)
			assert.Equal(t, 20, limit)
			return []*domain.SessionSet{
				{ID: uuid.New(), Weight: floatPtr(60.0), Reps: intPtr(10)},
				{ID: uuid.New(), Weight: floatPtr(65.0), Reps: intPtr(8)},
			}, false, nil
		},
	}

	handler := NewProgressHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/exercises/"+exerciseID.String()+"/history", nil)
	req = setURLParam(req, "id", exerciseID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.ExerciseHistory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	items, ok := resp["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 2)
	assert.Equal(t, false, resp["has_more"])
}

func TestProgressHandler_ExerciseHistory_InvalidUUID(t *testing.T) {
	mockUC := &MockProgressUseCase{}
	handler := NewProgressHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/exercises/not-a-uuid/history", nil)
	req = setURLParam(req, "id", "not-a-uuid")
	req = req.WithContext(contextWithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()

	handler.ExerciseHistory(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProgressHandler_ExerciseHistory_WithPagination(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockUC := &MockProgressUseCase{
		ExerciseHistoryFunc: func(ctx context.Context, uid, eid uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error) {
			assert.Equal(t, "test-cursor", cursor)
			assert.Equal(t, 10, limit)
			return []*domain.SessionSet{
				{ID: uuid.New(), Weight: floatPtr(60.0), Reps: intPtr(10)},
			}, false, nil
		},
	}

	handler := NewProgressHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/exercises/"+exerciseID.String()+"/history?cursor=test-cursor&limit=10", nil)
	req = setURLParam(req, "id", exerciseID.String())
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.ExerciseHistory(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProgressHandler_Summary_Success(t *testing.T) {
	userID := uuid.New()

	mockUC := &MockProgressUseCase{
		SummaryFunc: func(ctx context.Context, uid uuid.UUID) (*repository.ProgressSummary, error) {
			assert.Equal(t, userID, uid)
			return &repository.ProgressSummary{
				TotalSessions:      10,
				TotalWorkouts:      5,
				TotalExercises:     20,
				TotalTime:          3600,
				AvgSessionDuration: 1800,
			}, nil
		},
	}

	handler := NewProgressHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/summary", nil)
	req = req.WithContext(contextWithUserID(req.Context(), userID))
	w := httptest.NewRecorder()

	handler.Summary(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp repository.ProgressSummary
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 10, resp.TotalSessions)
	assert.Equal(t, 5, resp.TotalWorkouts)
	assert.Equal(t, 20, resp.TotalExercises)
	assert.Equal(t, 3600, resp.TotalTime)
	assert.Equal(t, 1800, resp.AvgSessionDuration)
}

func TestProgressHandler_Summary_Unauthorized(t *testing.T) {
	mockUC := &MockProgressUseCase{}
	handler := NewProgressHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/progress/summary", nil)
	w := httptest.NewRecorder()

	handler.Summary(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Helper functions for tests
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int          { return &i }
