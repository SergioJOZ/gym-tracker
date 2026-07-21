package testutil

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
)

// NewUser creates a new User with random data for testing.
func NewUser() *domain.User {
	return &domain.User{
		ID:        uuid.New(),
		Email:     fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewExercise creates a new Exercise with random data for testing.
func NewExercise() *domain.Exercise {
	return &domain.Exercise{
		ID:                 uuid.New(),
		NameByLang:         map[string]string{"en": fmt.Sprintf("Exercise-%s", uuid.New().String()[:8])},
		DescriptionsByLang: map[string]string{"en": "Test exercise description"},
		MuscleGroup:        "chest",
		Equipment:          "barbell",
		Difficulty:         "intermediate",
		Category:           "strength",
	}
}

// NewTemplate creates a new WorkoutTemplate with random data and slots for testing.
func NewTemplate(userID uuid.UUID) *domain.WorkoutTemplate {
	exerciseID := uuid.New()
	return &domain.WorkoutTemplate{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        fmt.Sprintf("Template-%s", uuid.New().String()[:8]),
		Description: "Test template description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Slots: []*domain.TemplateSlot{
			{
				ID:           uuid.New(),
				TemplateID:   uuid.New(), // Will be set by repository
				ExerciseID:   exerciseID,
				Order:        0,
				TargetSets:   3,
				TargetReps:   10,
				TargetWeight: floatPtr(60.0),
			},
			{
				ID:           uuid.New(),
				TemplateID:   uuid.New(), // Will be set by repository
				ExerciseID:   uuid.New(),
				Order:        1,
				TargetSets:   4,
				TargetReps:   8,
				TargetWeight: floatPtr(80.0),
			},
		},
	}
}

// NewSession creates a new WorkoutSession with random data, exercises, and sets for testing.
func NewSession(userID uuid.UUID) *domain.WorkoutSession {
	now := time.Now()
	end := now.Add(1 * time.Hour)
	return &domain.WorkoutSession{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      fmt.Sprintf("Session-%s", uuid.New().String()[:8]),
		Notes:     "Test session notes",
		StartAt:   now,
		EndAt:     &end,
		CreatedAt: now,
		UpdatedAt: now,
		Exercises: []*domain.SessionExercise{
			{
				ID:         uuid.New(),
				SessionID:  uuid.New(), // Will be set by repository
				ExerciseID: uuid.New(),
				Order:      0,
				Notes:      "First exercise",
				Sets: []*domain.SessionSet{
					{
						ID:               uuid.New(),
						SessionExerciseID: uuid.New(), // Will be set by repository
						Order:            0,
						Weight:           floatPtr(60.0),
						Reps:             intPtr(10),
					},
					{
						ID:               uuid.New(),
						SessionExerciseID: uuid.New(), // Will be set by repository
						Order:            1,
						Weight:           floatPtr(65.0),
						Reps:             intPtr(8),
					},
				},
			},
			{
				ID:         uuid.New(),
				SessionID:  uuid.New(), // Will be set by repository
				ExerciseID: uuid.New(),
				Order:      1,
				Notes:      "Second exercise",
				Sets: []*domain.SessionSet{
					{
						ID:               uuid.New(),
						SessionExerciseID: uuid.New(), // Will be set by repository
						Order:            0,
						Weight:           floatPtr(80.0),
						Reps:             intPtr(6),
					},
				},
			},
		},
	}
}

// NewPersonalRecord creates a new PersonalRecord with random data for testing.
func NewPersonalRecord(userID, exerciseID uuid.UUID) *domain.PersonalRecord {
	return &domain.PersonalRecord{
		ID:         uuid.New(),
		UserID:     userID,
		ExerciseID: exerciseID,
		MaxWeight:  floatPtr(100.0),
		MaxReps:    intPtr(5),
		MaxVolume:  floatPtr(500.0),
		UpdatedAt:  time.Now(),
	}
}

// Helper functions
func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int          { return &i }
