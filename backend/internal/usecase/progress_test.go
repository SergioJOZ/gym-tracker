package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProgressRepository is a mock implementation of ProgressRepository for testing.
type MockProgressRepository struct {
	ExerciseHistoryFunc func(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error)
	SummaryFunc         func(ctx context.Context, userID uuid.UUID) (*repository.ProgressSummary, error)
}

func (m *MockProgressRepository) ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error) {
	return m.ExerciseHistoryFunc(ctx, userID, exerciseID, cursor, limit)
}

func (m *MockProgressRepository) Summary(ctx context.Context, userID uuid.UUID) (*repository.ProgressSummary, error) {
	return m.SummaryFunc(ctx, userID)
}

func TestProgressUseCase_ListPRs(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockPRRepo := &MockPersonalRecordRepository{
		FindByUserFunc: func(ctx context.Context, uid uuid.UUID) ([]*domain.PersonalRecord, error) {
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

	uc := NewProgressUseCase(nil, mockPRRepo)

	prs, err := uc.ListPRs(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, prs, 1)
	assert.Equal(t, userID, prs[0].UserID)
	assert.Equal(t, exerciseID, prs[0].ExerciseID)
}

func TestProgressUseCase_ExerciseHistory(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockProgressRepo := &MockProgressRepository{
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

	uc := NewProgressUseCase(mockProgressRepo, nil)

	sets, hasMore, err := uc.ExerciseHistory(context.Background(), userID, exerciseID, "", 20)
	require.NoError(t, err)
	assert.Len(t, sets, 2)
	assert.False(t, hasMore)
}

func TestProgressUseCase_Summary(t *testing.T) {
	userID := uuid.New()

	mockProgressRepo := &MockProgressRepository{
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

	uc := NewProgressUseCase(mockProgressRepo, nil)

	summary, err := uc.Summary(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 10, summary.TotalSessions)
	assert.Equal(t, 5, summary.TotalWorkouts)
	assert.Equal(t, 20, summary.TotalExercises)
	assert.Equal(t, 3600, summary.TotalTime)
	assert.Equal(t, 1800, summary.AvgSessionDuration)
}
