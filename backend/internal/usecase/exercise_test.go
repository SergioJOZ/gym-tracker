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

// MockExerciseRepository is a mock implementation of ExerciseRepository for testing.
type MockExerciseRepository struct {
	ListFunc      func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error)
	GetByIDFunc   func(ctx context.Context, id uuid.UUID) (*domain.Exercise, error)
	BulkUpsertFunc func(ctx context.Context, exercises []*domain.Exercise) error
	ExistsFunc    func(ctx context.Context, id uuid.UUID) (bool, error)
}

func (m *MockExerciseRepository) List(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
	return m.ListFunc(ctx, filter)
}

func (m *MockExerciseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *MockExerciseRepository) BulkUpsert(ctx context.Context, exercises []*domain.Exercise) error {
	return m.BulkUpsertFunc(ctx, exercises)
}

func (m *MockExerciseRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, id)
	}
	return false, nil
}

func TestExerciseUseCase_List_Success(t *testing.T) {
	exercises := []*domain.Exercise{
		{ID: uuid.New(), Name: "Bench Press", MuscleGroup: "chest", Difficulty: "intermediate"},
		{ID: uuid.New(), Name: "Squat", MuscleGroup: "legs", Difficulty: "intermediate"},
	}

	mockRepo := &MockExerciseRepository{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			assert.Equal(t, "chest", filter.MuscleGroup)
			assert.Equal(t, 20, filter.Limit)
			return exercises, false, nil
		},
	}

	uc := NewExerciseUseCase(mockRepo)

	filter := repository.ExerciseFilter{
		MuscleGroup: "chest",
		Limit:       20,
	}

	results, hasMore, err := uc.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.False(t, hasMore)
	assert.Equal(t, "Bench Press", results[0].Name)
}

func TestExerciseUseCase_List_Empty(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			return []*domain.Exercise{}, false, nil
		},
	}

	uc := NewExerciseUseCase(mockRepo)

	results, hasMore, err := uc.List(context.Background(), repository.ExerciseFilter{Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.False(t, hasMore)
}

func TestExerciseUseCase_List_DefaultLimit(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			// Verify default limit is applied
			assert.Equal(t, 20, filter.Limit)
			return []*domain.Exercise{}, false, nil
		},
	}

	uc := NewExerciseUseCase(mockRepo)

	// Limit 0 should default to 20
	_, _, err := uc.List(context.Background(), repository.ExerciseFilter{})
	require.NoError(t, err)
}

func TestExerciseUseCase_GetByID_Success(t *testing.T) {
	exerciseID := uuid.New()
	exercise := &domain.Exercise{
		ID:          exerciseID,
		Name:        "Bench Press",
		Description: "Compound chest exercise",
		MuscleGroup: "chest",
		Equipment:   "barbell",
		Difficulty:  "intermediate",
		Category:    "strength",
	}

	mockRepo := &MockExerciseRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
			assert.Equal(t, exerciseID, id)
			return exercise, nil
		},
	}

	uc := NewExerciseUseCase(mockRepo)

	result, err := uc.GetByID(context.Background(), exerciseID)
	require.NoError(t, err)
	assert.Equal(t, exerciseID, result.ID)
	assert.Equal(t, "Bench Press", result.Name)
	assert.Equal(t, "chest", result.MuscleGroup)
}

func TestExerciseUseCase_GetByID_NotFound(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := NewExerciseUseCase(mockRepo)

	result, err := uc.GetByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)
}

func TestExerciseUseCase_List_WithSearch(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			assert.Equal(t, "bench", filter.Search)
			return []*domain.Exercise{
				{ID: uuid.New(), Name: "Bench Press", MuscleGroup: "chest", Difficulty: "beginner"},
			}, false, nil
		},
	}

	uc := NewExerciseUseCase(mockRepo)

	filter := repository.ExerciseFilter{
		Search: "bench",
		Limit:  10,
	}

	results, _, err := uc.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Bench Press", results[0].Name)
}

func TestExerciseUseCase_List_WithPagination(t *testing.T) {
	mockRepo := &MockExerciseRepository{
		ListFunc: func(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
			assert.Equal(t, "some-cursor", filter.Cursor)
			assert.Equal(t, 5, filter.Limit)
			return []*domain.Exercise{
				{ID: uuid.New(), Name: "Exercise 3", MuscleGroup: "chest", Difficulty: "beginner"},
			}, false, nil
		},
	}

	uc := NewExerciseUseCase(mockRepo)

	filter := repository.ExerciseFilter{
		Cursor: "some-cursor",
		Limit:  5,
	}

	results, hasMore, err := uc.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.False(t, hasMore)
}
