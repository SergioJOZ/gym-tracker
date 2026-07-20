package domain

import (
	"github.com/google/uuid"
)

// Valid difficulty levels for exercises.
var validDifficulties = map[string]bool{
	"beginner":     true,
	"intermediate": true,
	"advanced":     true,
}

// Exercise represents an exercise in the catalog.
type Exercise struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	MuscleGroup   string    `json:"muscle_group"`
	Equipment     string    `json:"equipment"`
	Difficulty    string    `json:"difficulty"`
	Category      string    `json:"category"`
	GIFPath       string    `json:"gif_path"`
	ThumbnailPath string    `json:"thumbnail_path"`
}

// Validate checks that required fields are present and sets defaults.
// Returns a validation AppError if any required field is missing.
func (e *Exercise) Validate() error {
	var details []FieldError

	if e.Name == "" {
		details = append(details, FieldError{Field: "name", Message: "is required"})
	}

	if e.MuscleGroup == "" {
		details = append(details, FieldError{Field: "muscle_group", Message: "is required"})
	}

	// Set default difficulty if empty
	if e.Difficulty == "" {
		e.Difficulty = "beginner"
	}

	// Validate difficulty value
	if !validDifficulties[e.Difficulty] {
		details = append(details, FieldError{Field: "difficulty", Message: "must be one of: beginner, intermediate, advanced"})
	}

	if len(details) > 0 {
		return NewValidationError("exercise validation failed", details)
	}

	return nil
}
