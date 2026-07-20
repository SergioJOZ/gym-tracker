package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WorkoutSession represents a completed or in-progress workout session.
type WorkoutSession struct {
	ID         uuid.UUID         `json:"id"`
	UserID     uuid.UUID         `json:"user_id"`
	TemplateID *uuid.UUID        `json:"template_id,omitempty"`
	Name       string            `json:"name"`
	Notes      string            `json:"notes"`
	StartAt    time.Time         `json:"start_at"`
	EndAt      *time.Time        `json:"end_at,omitempty"`
	Exercises  []*SessionExercise `json:"exercises,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// SessionExercise represents a single exercise within a workout session.
type SessionExercise struct {
	ID         uuid.UUID    `json:"id"`
	SessionID  uuid.UUID    `json:"session_id"`
	ExerciseID uuid.UUID    `json:"exercise_id"`
	Order      int          `json:"order"`
	Notes      string       `json:"notes"`
	Sets       []*SessionSet `json:"sets,omitempty"`
}

// SessionSet represents a single set within a session exercise.
type SessionSet struct {
	ID               uuid.UUID  `json:"id"`
	SessionExerciseID uuid.UUID `json:"session_exercise_id"`
	Order            int        `json:"order"`
	Weight           *float64   `json:"weight,omitempty"`
	Reps             *int       `json:"reps,omitempty"`
	Duration         *int       `json:"duration,omitempty"` // duration in seconds
	RPE              *float64   `json:"rpe,omitempty"`
}

// PersonalRecord represents a user's personal record for an exercise.
type PersonalRecord struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	ExerciseID uuid.UUID `json:"exercise_id"`
	MaxWeight  *float64  `json:"max_weight,omitempty"`
	MaxReps    *int      `json:"max_reps,omitempty"`
	MaxVolume  *float64  `json:"max_volume,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate checks that the session has required fields and valid nested data.
func (s *WorkoutSession) Validate() error {
	var details []FieldError

	if s.Name == "" {
		details = append(details, FieldError{Field: "name", Message: "is required"})
	}

	if len(s.Exercises) == 0 {
		details = append(details, FieldError{Field: "exercises", Message: "at least one exercise is required"})
	}

	for i, ex := range s.Exercises {
		if ex.Order < 0 {
			details = append(details, FieldError{
				Field:   fmt.Sprintf("exercises[%d].order", i),
				Message: "must be >= 0",
			})
		}

		if len(ex.Sets) == 0 {
			details = append(details, FieldError{
				Field:   fmt.Sprintf("exercises[%d].sets", i),
				Message: "at least one set is required per exercise",
			})
		}

		for j, set := range ex.Sets {
			if set.Order < 0 {
				details = append(details, FieldError{
					Field:   fmt.Sprintf("exercises[%d].sets[%d].order", i, j),
					Message: "must be >= 0",
				})
			}
		}
	}

	if len(details) > 0 {
		return NewValidationError("session validation failed", details)
	}

	return nil
}

// Validate checks that the personal record has required fields and at least one PR value.
func (pr *PersonalRecord) Validate() error {
	var details []FieldError

	if pr.UserID == uuid.Nil {
		details = append(details, FieldError{Field: "user_id", Message: "is required"})
	}

	if pr.ExerciseID == uuid.Nil {
		details = append(details, FieldError{Field: "exercise_id", Message: "is required"})
	}

	hasValue := false
	if pr.MaxWeight != nil && *pr.MaxWeight > 0 {
		hasValue = true
	}
	if pr.MaxReps != nil && *pr.MaxReps > 0 {
		hasValue = true
	}
	if pr.MaxVolume != nil && *pr.MaxVolume > 0 {
		hasValue = true
	}

	if !hasValue {
		details = append(details, FieldError{
			Field:   "pr_values",
			Message: "at least one PR value (max_weight, max_reps, or max_volume) must be greater than 0",
		})
	}

	if len(details) > 0 {
		return NewValidationError("personal record validation failed", details)
	}

	return nil
}
