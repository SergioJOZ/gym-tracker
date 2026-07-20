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
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTemplateTest(t *testing.T) (*testutil.TestDB, *postgres.TemplateRepository) {
	t.Helper()
	tdb := testutil.NewTestDB(t)
	repo := postgres.NewTemplateRepository(tdb.DB)
	return tdb, repo
}

// seedUser inserts a test user and returns their ID.
func seedUser(t *testing.T, tdb *testutil.TestDB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := tdb.DB.Exec(`INSERT INTO users (id, email, password) VALUES ($1, $2, $3)`,
		id, "test-"+id.String()+"@example.com", "hashedpw")
	require.NoError(t, err)
	return id
}

// seedExercise inserts a test exercise and returns its ID.
func seedExercise(t *testing.T, tdb *testutil.TestDB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := tdb.DB.Exec(`INSERT INTO exercises (id, name, muscle_group) VALUES ($1, $2, $3)`,
		id, "Exercise-"+id.String()[:8], "chest")
	require.NoError(t, err)
	return id
}

func TestTemplateRepository_Create_And_FindByID(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	exerciseID := seedExercise(t, tdb)

	tmpl := &domain.WorkoutTemplate{
		UserID:      userID,
		Name:        "Push Day",
		Description: "Chest and shoulders",
		Slots: []*domain.TemplateSlot{
			{
				ExerciseID:   exerciseID,
				Order:        0,
				TargetSets:   4,
				TargetReps:   10,
				TargetWeight: floatPtr(80.0),
			},
		},
	}

	ctx := context.Background()
	err := repo.Create(ctx, tmpl)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tmpl.ID)
	assert.NotEqual(t, uuid.Nil, tmpl.Slots[0].ID)
	assert.NotEqual(t, uuid.Nil, tmpl.Slots[0].TemplateID)

	// FindByID
	found, err := repo.FindByID(ctx, userID, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, tmpl.ID, found.ID)
	assert.Equal(t, "Push Day", found.Name)
	assert.Equal(t, "Chest and shoulders", found.Description)
	assert.Len(t, found.Slots, 1)
	assert.Equal(t, exerciseID, found.Slots[0].ExerciseID)
	assert.Equal(t, 0, found.Slots[0].Order)
	assert.Equal(t, 4, found.Slots[0].TargetSets)
	assert.Equal(t, 10, found.Slots[0].TargetReps)
	require.NotNil(t, found.Slots[0].TargetWeight)
	assert.InDelta(t, 80.0, *found.Slots[0].TargetWeight, 0.01)
}

func TestTemplateRepository_FindByID_WrongUser(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	otherUserID := seedUser(t, tdb)

	tmpl := &domain.WorkoutTemplate{
		UserID: userID,
		Name:   "My Template",
	}

	ctx := context.Background()
	err := repo.Create(ctx, tmpl)
	require.NoError(t, err)

	// Other user should not see it
	_, err = repo.FindByID(ctx, otherUserID, tmpl.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTemplateRepository_FindByID_NotFound(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, userID, uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTemplateRepository_Update(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ex1 := seedExercise(t, tdb)
	ex2 := seedExercise(t, tdb)

	tmpl := &domain.WorkoutTemplate{
		UserID: userID,
		Name:   "Original",
		Slots: []*domain.TemplateSlot{
			{ExerciseID: ex1, Order: 0, TargetSets: 3, TargetReps: 10},
		},
	}

	ctx := context.Background()
	err := repo.Create(ctx, tmpl)
	require.NoError(t, err)

	// Update: change name, replace slots
	tmpl.Name = "Updated"
	tmpl.Description = "New description"
	tmpl.Slots = []*domain.TemplateSlot{
		{ExerciseID: ex1, Order: 0, TargetSets: 5, TargetReps: 5},
		{ExerciseID: ex2, Order: 1, TargetSets: 3, TargetReps: 12},
	}

	err = repo.Update(ctx, tmpl)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, userID, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", found.Name)
	assert.Equal(t, "New description", found.Description)
	assert.Len(t, found.Slots, 2)
	assert.Equal(t, 5, found.Slots[0].TargetSets)
	assert.Equal(t, ex2, found.Slots[1].ExerciseID)
}

func TestTemplateRepository_Delete(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)

	tmpl := &domain.WorkoutTemplate{
		UserID: userID,
		Name:   "To Delete",
		Slots: []*domain.TemplateSlot{
			{ExerciseID: seedExercise(t, tdb), Order: 0},
		},
	}

	ctx := context.Background()
	err := repo.Create(ctx, tmpl)
	require.NoError(t, err)

	err = repo.Delete(ctx, userID, tmpl.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, userID, tmpl.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTemplateRepository_Delete_WrongUser(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	otherUserID := seedUser(t, tdb)

	tmpl := &domain.WorkoutTemplate{
		UserID: userID,
		Name:   "Protected",
	}

	ctx := context.Background()
	err := repo.Create(ctx, tmpl)
	require.NoError(t, err)

	// Other user cannot delete
	err = repo.Delete(ctx, otherUserID, tmpl.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// Still exists for owner
	found, err := repo.FindByID(ctx, userID, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, "Protected", found.Name)
}

func TestTemplateRepository_List_Pagination(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ctx := context.Background()

	// Create 5 templates
	for i := 0; i < 5; i++ {
		tmpl := &domain.WorkoutTemplate{
			UserID: userID,
			Name:   "Template",
		}
		err := repo.Create(ctx, tmpl)
		require.NoError(t, err)
	}

	// List with limit 3
	templates, hasMore, err := repo.List(ctx, userID, repository.TemplateFilter{Limit: 3})
	require.NoError(t, err)
	assert.Len(t, templates, 3)
	assert.True(t, hasMore)

	// List next page using cursor
	templates2, hasMore2, err := repo.List(ctx, userID, repository.TemplateFilter{
		Cursor: encodeCreatedAtCursor(templates[2].CreatedAt, templates[2].ID),
		Limit:  3,
	})
	require.NoError(t, err)
	assert.Len(t, templates2, 2)
	assert.False(t, hasMore2)
}

func TestTemplateRepository_List_UserScoped(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	user1 := seedUser(t, tdb)
	user2 := seedUser(t, tdb)
	ctx := context.Background()

	// User1 creates 2 templates
	for i := 0; i < 2; i++ {
		err := repo.Create(ctx, &domain.WorkoutTemplate{UserID: user1, Name: "U1"})
		require.NoError(t, err)
	}
	// User2 creates 1 template
	err := repo.Create(ctx, &domain.WorkoutTemplate{UserID: user2, Name: "U2"})
	require.NoError(t, err)

	templates, hasMore, err := repo.List(ctx, user1, repository.TemplateFilter{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, templates, 2)
	assert.False(t, hasMore)

	templates2, _, err := repo.List(ctx, user2, repository.TemplateFilter{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, templates2, 1)
}

func TestTemplateRepository_List_Empty(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ctx := context.Background()

	templates, hasMore, err := repo.List(ctx, userID, repository.TemplateFilter{Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, templates)
	assert.False(t, hasMore)
}

func TestTemplateRepository_Create_NoSlots(t *testing.T) {
	tdb, repo := setupTemplateTest(t)
	defer tdb.Cleanup(t)

	userID := seedUser(t, tdb)
	ctx := context.Background()

	tmpl := &domain.WorkoutTemplate{
		UserID:      userID,
		Name:        "Empty Template",
		Description: "No exercises yet",
	}

	err := repo.Create(ctx, tmpl)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tmpl.ID)

	found, err := repo.FindByID(ctx, userID, tmpl.ID)
	require.NoError(t, err)
	assert.Equal(t, "Empty Template", found.Name)
	assert.Empty(t, found.Slots)
}

// floatPtr returns a pointer to a float64 value.
func floatPtr(f float64) *float64 { return &f }

// encodeCreatedAtCursor encodes a cursor from created_at timestamp and ID.
func encodeCreatedAtCursor(ts time.Time, id uuid.UUID) string {
	return cursor.Encode(ts.Format(time.RFC3339Nano), id.String())
}
