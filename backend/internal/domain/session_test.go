package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(i int) *int { return &i }

func TestWorkoutSession_Validate_Valid(t *testing.T) {
	s := &WorkoutSession{
		UserID:  uuid.New(),
		Name:    "Morning Push",
		StartAt: time.Now(),
		Exercises: []*SessionExercise{
			{
				ExerciseID: uuid.New(),
				Order:      0,
				Sets: []*SessionSet{
					{Order: 0, Weight: floatPtr(80), Reps: intPtr(10)},
				},
			},
		},
	}

	err := s.Validate()
	require.NoError(t, err)
}

func TestWorkoutSession_Validate_EmptyName(t *testing.T) {
	s := &WorkoutSession{
		UserID: uuid.New(),
		Name:   "",
		Exercises: []*SessionExercise{
			{
				ExerciseID: uuid.New(),
				Order:      0,
				Sets: []*SessionSet{
					{Order: 0, Reps: intPtr(10)},
				},
			},
		},
	}

	err := s.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)

	found := false
	for _, d := range appErr.Details {
		if d.Field == "name" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'name' field in validation details")
}

func TestWorkoutSession_Validate_NoExercises(t *testing.T) {
	s := &WorkoutSession{
		UserID: uuid.New(),
		Name:   "Empty Session",
	}

	err := s.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)

	found := false
	for _, d := range appErr.Details {
		if d.Field == "exercises" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'exercises' field in validation details")
}

func TestWorkoutSession_Validate_ExerciseWithNoSets(t *testing.T) {
	s := &WorkoutSession{
		UserID: uuid.New(),
		Name:   "Bad Session",
		Exercises: []*SessionExercise{
			{
				ExerciseID: uuid.New(),
				Order:      0,
				Sets:       nil, // no sets
			},
		},
	}

	err := s.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)

	found := false
	for _, d := range appErr.Details {
		if d.Field == "exercises[0].sets" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'exercises[0].sets' field in validation details")
}

func TestWorkoutSession_Validate_NegativeExerciseOrder(t *testing.T) {
	s := &WorkoutSession{
		UserID: uuid.New(),
		Name:   "Bad Order",
		Exercises: []*SessionExercise{
			{
				ExerciseID: uuid.New(),
				Order:      -1,
				Sets: []*SessionSet{
					{Order: 0, Reps: intPtr(5)},
				},
			},
		},
	}

	err := s.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)

	found := false
	for _, d := range appErr.Details {
		if d.Field == "exercises[0].order" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'exercises[0].order' field in validation details")
}

func TestWorkoutSession_Validate_MultipleErrors(t *testing.T) {
	s := &WorkoutSession{
		UserID: uuid.New(),
		Name:   "",
		// No exercises
	}

	err := s.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(appErr.Details), 2, "expected at least 2 field errors")
}

func TestPersonalRecord_Validate_Valid(t *testing.T) {
	pr := &PersonalRecord{
		UserID:     uuid.New(),
		ExerciseID: uuid.New(),
		MaxWeight:  floatPtr(100.0),
		MaxReps:    intPtr(5),
	}

	err := pr.Validate()
	require.NoError(t, err)
}

func TestPersonalRecord_Validate_NoUserID(t *testing.T) {
	pr := &PersonalRecord{
		ExerciseID: uuid.New(),
		MaxWeight:  floatPtr(100.0),
	}

	err := pr.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)

	found := false
	for _, d := range appErr.Details {
		if d.Field == "user_id" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'user_id' field in validation details")
}

func TestPersonalRecord_Validate_NoExerciseID(t *testing.T) {
	pr := &PersonalRecord{
		UserID:    uuid.New(),
		MaxWeight: floatPtr(100.0),
	}

	err := pr.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)

	found := false
	for _, d := range appErr.Details {
		if d.Field == "exercise_id" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'exercise_id' field in validation details")
}

func TestPersonalRecord_Validate_NoPRValues(t *testing.T) {
	pr := &PersonalRecord{
		UserID:     uuid.New(),
		ExerciseID: uuid.New(),
		// All PR values are nil/zero
	}

	err := pr.Validate()
	require.Error(t, err)

	appErr, ok := err.(*AppError)
	require.True(t, ok)

	found := false
	for _, d := range appErr.Details {
		if d.Field == "pr_values" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'pr_values' field in validation details")
}

func TestPersonalRecord_Validate_OnlyMaxVolume(t *testing.T) {
	pr := &PersonalRecord{
		UserID:     uuid.New(),
		ExerciseID: uuid.New(),
		MaxVolume:  floatPtr(500.0),
	}

	err := pr.Validate()
	require.NoError(t, err)
}

func TestPersonalRecord_Validate_OnlyMaxReps(t *testing.T) {
	pr := &PersonalRecord{
		UserID:     uuid.New(),
		ExerciseID: uuid.New(),
		MaxReps:    intPtr(20),
	}

	err := pr.Validate()
	require.NoError(t, err)
}
