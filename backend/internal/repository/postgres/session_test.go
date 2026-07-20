package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSessionTest(t *testing.T) (*testutil.TestDB, *postgres.SessionRepository) {
	t.Helper()
	tdb := testutil.NewTestDB(t)
	repo := postgres.NewSessionRepository(tdb.DB)
	return tdb, repo
}

func TestSessionRepository_Create_And_FindByID(t *testing.T) {
	tdb, repo := setupSessionTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)

	session := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "Morning Push",
		Notes:   "Felt strong",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Notes:      "Bench press",
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(80.0), Reps: intPtr(10)},
					{Order: 1, Weight: floatPtr(85.0), Reps: intPtr(8)},
				},
			},
		},
	}

	ctx := context.Background()
	err := repo.Create(ctx, session)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
	assert.NotEqual(t, uuid.Nil, session.Exercises[0].ID)
	assert.NotEqual(t, uuid.Nil, session.Exercises[0].Sets[0].ID)

	// FindByID
	found, err := repo.FindByID(ctx, userID, session.ID)
	require.NoError(t, err)
	assert.Equal(t, session.ID, found.ID)
	assert.Equal(t, "Morning Push", found.Name)
	assert.Equal(t, "Felt strong", found.Notes)
	require.Len(t, found.Exercises, 1)
	assert.Equal(t, exerciseID, found.Exercises[0].ExerciseID)
	assert.Equal(t, "Bench press", found.Exercises[0].Notes)
	require.Len(t, found.Exercises[0].Sets, 2)
	assert.InDelta(t, 80.0, *found.Exercises[0].Sets[0].Weight, 0.01)
	assert.Equal(t, 10, *found.Exercises[0].Sets[0].Reps)
}

func TestSessionRepository_FindByID_WrongUser(t *testing.T) {
	tdb, repo := setupSessionTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	otherUserID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)

	session := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "My Session",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets:       []*domain.SessionSet{{Order: 0, Reps: intPtr(10)}},
			},
		},
	}

	ctx := context.Background()
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Other user should not see it
	_, err = repo.FindByID(ctx, otherUserID, session.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestSessionRepository_Update(t *testing.T) {
	tdb, repo := setupSessionTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ex1 := seedExercise(t, tdb)
	ex2 := seedExercise(t, tdb)

	session := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "Original",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: ex1,
				Order:      0,
				Sets:       []*domain.SessionSet{{Order: 0, Weight: floatPtr(60.0), Reps: intPtr(12)}},
			},
		},
	}

	ctx := context.Background()
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Update: change name, replace exercises and sets
	session.Name = "Updated Session"
	session.Notes = "New notes"
	session.Exercises = []*domain.SessionExercise{
		{
			ExerciseID: ex1,
			Order:      0,
			Sets: []*domain.SessionSet{
				{Order: 0, Weight: floatPtr(70.0), Reps: intPtr(10)},
				{Order: 1, Weight: floatPtr(75.0), Reps: intPtr(8)},
			},
		},
		{
			ExerciseID: ex2,
			Order:      1,
			Sets:       []*domain.SessionSet{{Order: 0, Reps: intPtr(15)}},
		},
	}

	err = repo.Update(ctx, session)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, userID, session.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Session", found.Name)
	assert.Equal(t, "New notes", found.Notes)
	require.Len(t, found.Exercises, 2)
	assert.Equal(t, ex2, found.Exercises[1].ExerciseID)
	require.Len(t, found.Exercises[0].Sets, 2)
}

func TestSessionRepository_Delete(t *testing.T) {
	tdb, repo := setupSessionTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)

	session := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "To Delete",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets:       []*domain.SessionSet{{Order: 0, Reps: intPtr(10)}},
			},
		},
	}

	ctx := context.Background()
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	err = repo.Delete(ctx, userID, session.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, userID, session.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestSessionRepository_List_Pagination(t *testing.T) {
	tdb, repo := setupSessionTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	// Create 5 sessions
	for i := 0; i < 5; i++ {
		session := &domain.WorkoutSession{
			UserID:  userID,
			Name:    "Session",
			StartAt: time.Now(),
			Exercises: []*domain.SessionExercise{
				{
					ExerciseID: exerciseID,
					Order:      0,
					Sets:       []*domain.SessionSet{{Order: 0, Reps: intPtr(10)}},
				},
			},
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// List with limit 3
	sessions, hasMore, err := repo.List(ctx, userID, repository.SessionFilter{Limit: 3})
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
	assert.True(t, hasMore)
}

func TestSessionRepository_List_UserScoped(t *testing.T) {
	tdb, repo := setupSessionTest(t)
	defer tdb.Cleanup(t)

	user1 := seedUser(t, tdb)
	user2 := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)
	ctx := context.Background()

	// User1 creates 2 sessions
	for i := 0; i < 2; i++ {
		session := &domain.WorkoutSession{
			UserID:  user1,
			Name:    "U1 Session",
			StartAt: time.Now(),
			Exercises: []*domain.SessionExercise{
				{
					ExerciseID: exerciseID,
					Order:      0,
					Sets:       []*domain.SessionSet{{Order: 0, Reps: intPtr(10)}},
				},
			},
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// User2 creates 1 session
	session := &domain.WorkoutSession{
		UserID:  user2,
		Name:    "U2 Session",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets:       []*domain.SessionSet{{Order: 0, Reps: intPtr(10)}},
			},
		},
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	sessions, hasMore, err := repo.List(ctx, user1, repository.SessionFilter{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.False(t, hasMore)

	sessions2, _, err := repo.List(ctx, user2, repository.SessionFilter{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, sessions2, 1)
}

func intPtr(i int) *int { return &i }
