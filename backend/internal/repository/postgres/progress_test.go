package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProgressTest(t *testing.T) (*testutil.TestDB, *postgres.ProgressRepository) {
	t.Helper()
	tdb := testutil.NewTestDB(t)
	repo := postgres.NewProgressRepository(tdb.DB)
	return tdb, repo
}

func TestProgressRepository_ExerciseHistory(t *testing.T) {
	tdb, repo := setupProgressTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	// Create a session repo to create test data
	sessionRepo := postgres.NewSessionRepository(tdb.DB)

	// Create 3 sessions with the same exercise
	for i := 0; i < 3; i++ {
		session := &domain.WorkoutSession{
			UserID:  userID,
			Name:    "Session",
			StartAt: time.Now().Add(time.Duration(i) * time.Hour),
			Exercises: []*domain.SessionExercise{
				{
					ExerciseID: exerciseID,
					Order:      0,
					Sets: []*domain.SessionSet{
						{Order: 0, Weight: floatPtr(60.0 + float64(i)*5), Reps: intPtr(10 + i)},
					},
				},
			},
		}
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Test ExerciseHistory
	sets, hasMore, err := repo.ExerciseHistory(ctx, userID, exerciseID, "", 10)
	require.NoError(t, err)
	assert.Len(t, sets, 3)
	assert.False(t, hasMore)

	// Verify ordering (most recent first)
	assert.InDelta(t, 70.0, *sets[0].Weight, 0.01) // Latest session
	assert.InDelta(t, 60.0, *sets[2].Weight, 0.01) // Oldest session
}

func TestProgressRepository_ExerciseHistory_Pagination(t *testing.T) {
	tdb, repo := setupProgressTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	sessionRepo := postgres.NewSessionRepository(tdb.DB)

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		session := &domain.WorkoutSession{
			UserID:  userID,
			Name:    "Session",
			StartAt: time.Now().Add(time.Duration(i) * time.Hour),
			Exercises: []*domain.SessionExercise{
				{
					ExerciseID: exerciseID,
					Order:      0,
					Sets: []*domain.SessionSet{
						{Order: 0, Weight: floatPtr(60.0), Reps: intPtr(10)},
					},
				},
			},
		}
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Get first page with limit 2
	sets, hasMore, err := repo.ExerciseHistory(ctx, userID, exerciseID, "", 2)
	require.NoError(t, err)
	assert.Len(t, sets, 2)
	assert.True(t, hasMore)
}

func TestProgressRepository_ExerciseHistory_UserScoped(t *testing.T) {
	tdb, repo := setupProgressTest(t)
	defer tdb.Cleanup(t)

	user1 := seedUser(t, tdb)
	user2 := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	sessionRepo := postgres.NewSessionRepository(tdb.DB)

	// User1 creates 2 sessions
	for i := 0; i < 2; i++ {
		session := &domain.WorkoutSession{
			UserID:  user1,
			Name:    "Session",
			StartAt: time.Now(),
			Exercises: []*domain.SessionExercise{
				{
					ExerciseID: exerciseID,
					Order:      0,
					Sets: []*domain.SessionSet{
						{Order: 0, Weight: floatPtr(60.0), Reps: intPtr(10)},
					},
				},
			},
		}
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)
	}

	// User2 creates 1 session
	session := &domain.WorkoutSession{
		UserID:  user2,
		Name:    "Session",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(70.0), Reps: intPtr(8)},
				},
			},
		},
	}
	err := sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// User1 should only see their own sets
	sets, _, err := repo.ExerciseHistory(ctx, user1, exerciseID, "", 10)
	require.NoError(t, err)
	assert.Len(t, sets, 2)

	// User2 should only see their own sets
	sets2, _, err := repo.ExerciseHistory(ctx, user2, exerciseID, "", 10)
	require.NoError(t, err)
	assert.Len(t, sets2, 1)
}

func TestProgressRepository_Summary(t *testing.T) {
	tdb, repo := setupProgressTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	sessionRepo := postgres.NewSessionRepository(tdb.DB)

	// Create 2 sessions with different durations
	start1 := time.Now().Add(-2 * time.Hour)
	end1 := time.Now().Add(-1 * time.Hour)
	session1 := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "Session 1",
		StartAt: start1,
		EndAt:   &end1,
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(60.0), Reps: intPtr(10)},
					{Order: 1, Weight: floatPtr(65.0), Reps: intPtr(8)},
				},
			},
		},
	}
	err := sessionRepo.Create(ctx, session1)
	require.NoError(t, err)

	start2 := time.Now().Add(-1 * time.Hour)
	end2 := time.Now()
	session2 := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "Session 2",
		StartAt: start2,
		EndAt:   &end2,
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(70.0), Reps: intPtr(6)},
				},
			},
		},
	}
	err = sessionRepo.Create(ctx, session2)
	require.NoError(t, err)

	// Get summary
	summary, err := repo.Summary(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, summary.TotalSessions)
	assert.Equal(t, 2, summary.TotalWorkouts)
	assert.Equal(t, 2, summary.TotalExercises) // 2 unique exercises (same exercise ID)
	assert.Greater(t, summary.TotalTime, 0)
}

func TestProgressRepository_Summary_Empty(t *testing.T) {
	tdb, repo := setupProgressTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ctx := context.Background()

	// Get summary for user with no sessions
	summary, err := repo.Summary(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, summary.TotalSessions)
	assert.Equal(t, 0, summary.TotalWorkouts)
	assert.Equal(t, 0, summary.TotalExercises)
	assert.Equal(t, 0, summary.TotalTime)
	assert.Equal(t, 0, summary.AvgSessionDuration)
}
