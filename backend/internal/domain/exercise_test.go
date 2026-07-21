package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestExercise_Validate_Valid(t *testing.T) {
	ex := &Exercise{
		ID:                 uuid.New(),
		NameByLang:         map[string]string{"en": "Bench Press"},
		DescriptionsByLang: map[string]string{"en": "A compound exercise targeting the chest."},
		MuscleGroup:        "chest",
		Equipment:          "barbell",
		Difficulty:         "intermediate",
		Category:           "strength",
		GIFPath:            "/gifs/bench_press.gif",
		ThumbnailPath:      "/thumbnails/bench_press.jpg",
	}

	if err := ex.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestExercise_Validate_EmptyName(t *testing.T) {
	ex := &Exercise{
		ID:          uuid.New(),
		NameByLang:  map[string]string{"en": ""},
		MuscleGroup: "chest",
	}

	err := ex.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %q", appErr.Code)
	}
	if len(appErr.Details) == 0 {
		t.Fatal("expected field details in validation error")
	}

	found := false
	for _, d := range appErr.Details {
		if d.Field == "name_by_lang" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'name_by_lang' field in validation details")
	}
}

func TestExercise_Validate_EmptyMuscleGroup(t *testing.T) {
	ex := &Exercise{
		ID:          uuid.New(),
		NameByLang:  map[string]string{"en": "Bench Press"},
		MuscleGroup: "",
	}

	err := ex.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty muscle_group")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}

	found := false
	for _, d := range appErr.Details {
		if d.Field == "muscle_group" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'muscle_group' field in validation details")
	}
}

func TestExercise_Validate_MultipleErrors(t *testing.T) {
	ex := &Exercise{
		ID:          uuid.New(),
		MuscleGroup: "",
	}

	err := ex.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}

	if len(appErr.Details) < 2 {
		t.Errorf("expected at least 2 field errors, got %d", len(appErr.Details))
	}
}

func TestExercise_Validate_Defaults(t *testing.T) {
	ex := &Exercise{
		ID:          uuid.New(),
		NameByLang:  map[string]string{"en": "Push Up"},
		MuscleGroup: "chest",
	}

	// Validate should set defaults for optional fields
	if err := ex.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ex.Difficulty != "beginner" {
		t.Errorf("expected default difficulty 'beginner', got %q", ex.Difficulty)
	}
}

func TestExercise_Validate_InvalidDifficulty(t *testing.T) {
	ex := &Exercise{
		ID:          uuid.New(),
		NameByLang:  map[string]string{"en": "Push Up"},
		MuscleGroup: "chest",
		Difficulty:  "invalid_level",
	}

	err := ex.Validate()
	if err == nil {
		t.Fatal("expected validation error for invalid difficulty")
	}
}
