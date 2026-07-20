package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPersonalRecordTest(t *testing.T) (*testutil.TestDB, *postgres.PersonalRecordRepository) {
	t.Helper()
	tdb := testutil.NewTestDB(t)
	repo := postgres.NewPersonalRecordRepository(tdb.DB)
	return tdb, repo
}

func TestPersonalRecordRepository_Upsert_Insert(t *testing.T) {
	tdb, repo := setupPersonalRecordTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)

	pr := &domain.PersonalRecord{
		UserID:     userID,
		ExerciseID: exerciseID,
		MaxWeight:  floatPtr(100.0),
		MaxReps:    intPtr(5),
		MaxVolume:  floatPtr(500.0),
	}

	ctx := context.Background()
	err := repo.Upsert(ctx, pr)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, pr.ID)

	// Verify it was inserted
	found, err := repo.FindByUserAndExercise(ctx, userID, exerciseID)
	require.NoError(t, err)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, exerciseID, found.ExerciseID)
	assert.InDelta(t, 100.0, *found.MaxWeight, 0.01)
	assert.Equal(t, 5, *found.MaxReps)
	assert.InDelta(t, 500.0, *found.MaxVolume, 0.01)
}

func TestPersonalRecordRepository_Upsert_UpdateWithGreatest(t *testing.T) {
	tdb, repo := setupPersonalRecordTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	// Insert initial PR
	pr1 := &domain.PersonalRecord{
		UserID:     userID,
		ExerciseID: exerciseID,
		MaxWeight:  floatPtr(100.0),
		MaxReps:    intPtr(5),
	}
	err := repo.Upsert(ctx, pr1)
	require.NoError(t, err)

	// Upsert with higher weight but lower reps - should keep highest of each
	pr2 := &domain.PersonalRecord{
		UserID:     userID,
		ExerciseID: exerciseID,
		MaxWeight:  floatPtr(110.0), // higher
		MaxReps:    intPtr(3),       // lower
	}
	err = repo.Upsert(ctx, pr2)
	require.NoError(t, err)

	// Verify GREATEST was applied
	found, err := repo.FindByUserAndExercise(ctx, userID, exerciseID)
	require.NoError(t, err)
	assert.InDelta(t, 110.0, *found.MaxWeight, 0.01) // should be 110 (higher)
	assert.Equal(t, 5, *found.MaxReps)                // should be 5 (higher)
}

func TestPersonalRecordRepository_FindByUserAndExercise_NotFound(t *testing.T) {
	tdb, repo := setupPersonalRecordTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ctx := context.Background()

	_, err := repo.FindByUserAndExercise(ctx, userID, uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPersonalRecordRepository_FindByUser(t *testing.T) {
	tdb, repo := setupPersonalRecordTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ex1 := seedExercise(t, tdb)
	ex2 := seedExercise(t, tdb)
	ctx := context.Background()

	// Create 2 PRs for user
	pr1 := &domain.PersonalRecord{
		UserID:     userID,
		ExerciseID: ex1,
		MaxWeight:  floatPtr(100.0),
	}
	err := repo.Upsert(ctx, pr1)
	require.NoError(t, err)

	pr2 := &domain.PersonalRecord{
		UserID:     userID,
		ExerciseID: ex2,
		MaxWeight:  floatPtr(80.0),
	}
	err = repo.Upsert(ctx, pr2)
	require.NoError(t, err)

	// Find all PRs for user
	prs, err := repo.FindByUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, prs, 2)
}

func TestPersonalRecordRepository_RecalculateFromSessions(t *testing.T) {
	tdb, repo := setupPersonalRecordTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	// Create sessions with sets
	sessionRepo := postgres.NewSessionRepository(tdb.DB)

	session1 := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "Session 1",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(80.0), Reps: intPtr(10)},
					{Order: 1, Weight: floatPtr(90.0), Reps: intPtr(8)},
				},
			},
		},
	}
	err := sessionRepo.Create(ctx, session1)
	require.NoError(t, err)

	session2 := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "Session 2",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(100.0), Reps: intPtr(5)}, // max weight
					{Order: 1, Weight: floatPtr(70.0), Reps: intPtr(12)}, // max reps
				},
			},
		},
	}
	err = sessionRepo.Create(ctx, session2)
	require.NoError(t, err)

	// Recalculate PRs
	err = repo.RecalculateFromSessions(ctx, userID, []uuid.UUID{exerciseID})
	require.NoError(t, err)

	// Verify PRs were calculated correctly
	pr, err := repo.FindByUserAndExercise(ctx, userID, exerciseID)
	require.NoError(t, err)
	assert.InDelta(t, 100.0, *pr.MaxWeight, 0.01) // max weight from session2
	assert.Equal(t, 12, *pr.MaxReps)               // max reps from session2
	// max_volume = max(weight * reps) = 100*5=500 or 90*8=720 or 80*10=800 or 70*12=840
	assert.InDelta(t, 840.0, *pr.MaxVolume, 0.01)
}
