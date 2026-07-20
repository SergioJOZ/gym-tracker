package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
)

// MockProgressRepository is a mock implementation for testing interface compliance.
type MockProgressRepository struct {
	ExerciseHistoryFunc func(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error)
	SummaryFunc         func(ctx context.Context, userID uuid.UUID) (*ProgressSummary, error)
}

func (m *MockProgressRepository) ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error) {
	return m.ExerciseHistoryFunc(ctx, userID, exerciseID, cursor, limit)
}

func (m *MockProgressRepository) Summary(ctx context.Context, userID uuid.UUID) (*ProgressSummary, error) {
	return m.SummaryFunc(ctx, userID)
}

func TestProgressRepository_Interface(t *testing.T) {
	// Test that MockProgressRepository implements ProgressRepository interface
	var _ ProgressRepository = (*MockProgressRepository)(nil)
}

func TestProgressSummary_Struct(t *testing.T) {
	summary := &ProgressSummary{
		TotalSessions:     10,
		TotalWorkouts:     5,
		TotalExercises:    20,
		TotalTime:         3600,
		AvgSessionDuration: 1800,
	}

	assert.Equal(t, 10, summary.TotalSessions)
	assert.Equal(t, 5, summary.TotalWorkouts)
	assert.Equal(t, 20, summary.TotalExercises)
	assert.Equal(t, 3600, summary.TotalTime)
	assert.Equal(t, 1800, summary.AvgSessionDuration)
}

func TestProgressFilter_Struct(t *testing.T) {
	filter := ProgressFilter{
		Cursor: "test-cursor",
		Limit:  20,
	}

	assert.Equal(t, "test-cursor", filter.Cursor)
	assert.Equal(t, 20, filter.Limit)
}
