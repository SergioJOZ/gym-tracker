package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const maxTemplateSlots = 50

// WorkoutTemplate represents a reusable workout template owned by a user.
type WorkoutTemplate struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Slots       []*TemplateSlot `json:"slots,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TemplateSlot represents a single exercise slot within a workout template.
type TemplateSlot struct {
	ID             uuid.UUID  `json:"id"`
	TemplateID     uuid.UUID  `json:"template_id"`
	ExerciseID     uuid.UUID  `json:"exercise_id"`
	Order          int        `json:"order"`
	TargetSets     int        `json:"target_sets"`
	TargetReps     int        `json:"target_reps"`
	TargetWeight   *float64   `json:"target_weight,omitempty"`
	TargetDuration *int       `json:"target_duration,omitempty"` // duration in seconds
}

// Validate checks that the template has required fields and valid slots.
func (t *WorkoutTemplate) Validate() error {
	var details []FieldError

	if t.Name == "" {
		details = append(details, FieldError{Field: "name", Message: "is required"})
	}

	if len(t.Slots) > maxTemplateSlots {
		details = append(details, FieldError{
			Field:   "slots",
			Message: "maximum 50 slots allowed per template",
		})
	}

	for i, slot := range t.Slots {
		if slot.Order < 0 {
			details = append(details, FieldError{
				Field:   fmt.Sprintf("slots[%d].order", i),
				Message: "must be >= 0",
			})
		}
	}

	if len(details) > 0 {
		return NewValidationError("template validation failed", details)
	}

	return nil
}
