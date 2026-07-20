package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres/testutil"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExerciseRepo_BulkUpsert(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	exercises := []*domain.Exercise{
		{
			ID:          uuid.New(),
			Name:        "Bench Press",
			Description: "Compound chest exercise",
			MuscleGroup: "chest",
			Equipment:   "barbell",
			Difficulty:  "intermediate",
			Category:    "strength",
		},
		{
			ID:          uuid.New(),
			Name:        "Squat",
			Description: "Compound leg exercise",
			MuscleGroup: "legs",
			Equipment:   "barbell",
			Difficulty:  "intermediate",
			Category:    "strength",
		},
	}

	err := repo.BulkUpsert(context.Background(), exercises)
	require.NoError(t, err)

	// Verify they were inserted
	for _, ex := range exercises {
		found, err := repo.GetByID(context.Background(), ex.ID)
		require.NoError(t, err)
		assert.Equal(t, ex.Name, found.Name)
	}
}

func TestExerciseRepo_BulkUpsert_Update(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	ex := &domain.Exercise{
		ID:          uuid.New(),
		Name:        "Bench Press",
		Description: "Original description",
		MuscleGroup: "chest",
		Difficulty:  "beginner",
	}

	err := repo.BulkUpsert(context.Background(), []*domain.Exercise{ex})
	require.NoError(t, err)

	// Update the same exercise
	ex.Description = "Updated description"
	err = repo.BulkUpsert(context.Background(), []*domain.Exercise{ex})
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), ex.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", found.Description)
}

func TestExerciseRepo_GetByID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	ex := &domain.Exercise{
		ID:            uuid.New(),
		Name:          "Deadlift",
		Description:   "Compound back exercise",
		MuscleGroup:   "back",
		Equipment:     "barbell",
		Difficulty:    "advanced",
		Category:      "strength",
		GIFPath:       "/gifs/deadlift.gif",
		ThumbnailPath: "/thumbnails/deadlift.jpg",
	}

	err := repo.BulkUpsert(context.Background(), []*domain.Exercise{ex})
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), ex.ID)
	require.NoError(t, err)
	assert.Equal(t, ex.ID, found.ID)
	assert.Equal(t, "Deadlift", found.Name)
	assert.Equal(t, "back", found.MuscleGroup)
	assert.Equal(t, "barbell", found.Equipment)
	assert.Equal(t, "advanced", found.Difficulty)
	assert.Equal(t, "/gifs/deadlift.gif", found.GIFPath)
}

func TestExerciseRepo_GetByID_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestExerciseRepo_List_NoFilters(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	exercises := make([]*domain.Exercise, 5)
	for i := 0; i < 5; i++ {
		exercises[i] = &domain.Exercise{
			ID:          uuid.New(),
			Name:        "Exercise",
			MuscleGroup: "chest",
			Difficulty:  "beginner",
		}
	}

	err := repo.BulkUpsert(context.Background(), exercises)
	require.NoError(t, err)

	filter := repository.ExerciseFilter{Limit: 10}
	results, hasMore, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 5)
	assert.False(t, hasMore)
}

func TestExerciseRepo_List_Pagination(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	exercises := make([]*domain.Exercise, 5)
	for i := 0; i < 5; i++ {
		exercises[i] = &domain.Exercise{
			ID:          uuid.New(),
			Name:        "Exercise",
			MuscleGroup: "chest",
			Difficulty:  "beginner",
		}
	}

	err := repo.BulkUpsert(context.Background(), exercises)
	require.NoError(t, err)

	// Get first page of 2
	filter := repository.ExerciseFilter{Limit: 2}
	results, hasMore, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.True(t, hasMore)

	// Use cursor to get next page
	cursor := encodeExerciseCursor(results[len(results)-1])
	filter.Cursor = cursor
	results2, hasMore2, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results2, 2)
	assert.True(t, hasMore2)

	// Verify no overlap
	assert.NotEqual(t, results[0].ID, results2[0].ID)
	assert.NotEqual(t, results[1].ID, results2[1].ID)
}

func TestExerciseRepo_List_FilterByMuscleGroup(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	exercises := []*domain.Exercise{
		{ID: uuid.New(), Name: "Bench Press", MuscleGroup: "chest", Difficulty: "beginner"},
		{ID: uuid.New(), Name: "Squat", MuscleGroup: "legs", Difficulty: "beginner"},
		{ID: uuid.New(), Name: "Row", MuscleGroup: "back", Difficulty: "beginner"},
	}

	err := repo.BulkUpsert(context.Background(), exercises)
	require.NoError(t, err)

	filter := repository.ExerciseFilter{MuscleGroup: "chest", Limit: 10}
	results, hasMore, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.False(t, hasMore)
	assert.Equal(t, "Bench Press", results[0].Name)
}

func TestExerciseRepo_List_FilterByDifficulty(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	exercises := []*domain.Exercise{
		{ID: uuid.New(), Name: "Push Up", MuscleGroup: "chest", Difficulty: "beginner"},
		{ID: uuid.New(), Name: "Bench Press", MuscleGroup: "chest", Difficulty: "intermediate"},
		{ID: uuid.New(), Name: "Weighted Dip", MuscleGroup: "chest", Difficulty: "advanced"},
	}

	err := repo.BulkUpsert(context.Background(), exercises)
	require.NoError(t, err)

	filter := repository.ExerciseFilter{Difficulty: "advanced", Limit: 10}
	results, _, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Weighted Dip", results[0].Name)
}

func TestExerciseRepo_List_Search(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	exercises := []*domain.Exercise{
		{ID: uuid.New(), Name: "Bench Press", Description: "Compound chest exercise", MuscleGroup: "chest", Difficulty: "beginner"},
		{ID: uuid.New(), Name: "Squat", Description: "Compound leg exercise", MuscleGroup: "legs", Difficulty: "beginner"},
		{ID: uuid.New(), Name: "Bicep Curl", Description: "Isolation arm exercise", MuscleGroup: "arms", Difficulty: "beginner"},
	}

	err := repo.BulkUpsert(context.Background(), exercises)
	require.NoError(t, err)

	filter := repository.ExerciseFilter{Search: "bench", Limit: 10}
	results, _, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Bench Press", results[0].Name)
}

func TestExerciseRepo_List_SearchDescription(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	exercises := []*domain.Exercise{
		{ID: uuid.New(), Name: "Bench Press", Description: "Compound chest exercise", MuscleGroup: "chest", Difficulty: "beginner"},
		{ID: uuid.New(), Name: "Squat", Description: "Compound leg exercise", MuscleGroup: "legs", Difficulty: "beginner"},
	}

	err := repo.BulkUpsert(context.Background(), exercises)
	require.NoError(t, err)

	filter := repository.ExerciseFilter{Search: "chest", Limit: 10}
	results, _, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Bench Press", results[0].Name)
}

func TestExerciseRepo_List_EmptyResult(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewExerciseRepository(tdb.DB)

	filter := repository.ExerciseFilter{MuscleGroup: "nonexistent", Limit: 10}
	results, hasMore, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Empty(t, results)
	assert.False(t, hasMore)
}

// encodeExerciseCursor creates a cursor from an exercise's ID for testing pagination.
func encodeExerciseCursor(ex *domain.Exercise) string {
	return cursor.Encode(ex.ID.String())
}
