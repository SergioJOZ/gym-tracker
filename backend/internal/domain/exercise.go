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
	ID                 uuid.UUID          `json:"id"`
	NameByLang         map[string]string  `json:"name_by_lang"`
	DescriptionsByLang map[string]string  `json:"descriptions_by_lang"`
	MuscleGroup        string             `json:"muscle_group"`
	Equipment          string             `json:"equipment"`
	Difficulty         string             `json:"difficulty"`
	Category           string             `json:"category"`
	GIFPath            string             `json:"gif_path"`
	ThumbnailPath      string             `json:"thumbnail_path"`
}

// Validate checks that required fields are present and sets defaults.
// Returns a validation AppError if any required field is missing.
func (e *Exercise) Validate() error {
	var details []FieldError

	if len(e.NameByLang) == 0 || e.NameByLang["en"] == "" {
		details = append(details, FieldError{Field: "name_by_lang", Message: "at least 'en' entry is required"})
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
