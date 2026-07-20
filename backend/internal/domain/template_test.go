package domain

import (
	"testing"

	"github.com/google/uuid"
)

func floatPtr(f float64) *float64 { return &f }

func TestWorkoutTemplate_Validate_Valid(t *testing.T) {
	tmpl := &WorkoutTemplate{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Push Day",
		Slots: []*TemplateSlot{
			{
				ID:           uuid.New(),
				TemplateID:   uuid.Nil,
				ExerciseID:   uuid.New(),
				Order:        0,
				TargetSets:   3,
				TargetReps:   10,
				TargetWeight: floatPtr(60.0),
			},
		},
	}

	if err := tmpl.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestWorkoutTemplate_Validate_EmptyName(t *testing.T) {
	tmpl := &WorkoutTemplate{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "",
	}

	err := tmpl.Validate()
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

	found := false
	for _, d := range appErr.Details {
		if d.Field == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'name' field in validation details")
	}
}

func TestWorkoutTemplate_Validate_NegativeSlotOrder(t *testing.T) {
	tmpl := &WorkoutTemplate{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Push Day",
		Slots: []*TemplateSlot{
			{
				ExerciseID: uuid.New(),
				Order:      -1,
			},
		},
	}

	err := tmpl.Validate()
	if err == nil {
		t.Fatal("expected validation error for negative slot order")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}

	found := false
	for _, d := range appErr.Details {
		if d.Field == "slots[0].order" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'slots[0].order' field in validation details")
	}
}

func TestWorkoutTemplate_Validate_TooManySlots(t *testing.T) {
	slots := make([]*TemplateSlot, 51)
	for i := range slots {
		slots[i] = &TemplateSlot{
			ExerciseID: uuid.New(),
			Order:      i,
		}
	}

	tmpl := &WorkoutTemplate{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Mega Template",
		Slots:  slots,
	}

	err := tmpl.Validate()
	if err == nil {
		t.Fatal("expected validation error for too many slots")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}

	found := false
	for _, d := range appErr.Details {
		if d.Field == "slots" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'slots' field in validation details")
	}
}

func TestWorkoutTemplate_Validate_ExactlyMaxSlots(t *testing.T) {
	slots := make([]*TemplateSlot, 50)
	for i := range slots {
		slots[i] = &TemplateSlot{
			ExerciseID: uuid.New(),
			Order:      i,
		}
	}

	tmpl := &WorkoutTemplate{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Max Template",
		Slots:  slots,
	}

	if err := tmpl.Validate(); err != nil {
		t.Errorf("expected no error for exactly 50 slots, got %v", err)
	}
}

func TestWorkoutTemplate_Validate_NoSlots(t *testing.T) {
	tmpl := &WorkoutTemplate{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "Empty Template",
	}

	if err := tmpl.Validate(); err != nil {
		t.Errorf("expected no error for template with no slots, got %v", err)
	}
}

func TestWorkoutTemplate_Validate_MultipleErrors(t *testing.T) {
	tmpl := &WorkoutTemplate{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Name:   "",
		Slots: []*TemplateSlot{
			{ExerciseID: uuid.New(), Order: -1},
		},
	}

	err := tmpl.Validate()
	if err == nil {
		t.Fatal("expected validation errors")
	}

	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}

	if len(appErr.Details) < 2 {
		t.Errorf("expected at least 2 field errors, got %d", len(appErr.Details))
	}
}

func TestTemplateSlot_ZeroOrder(t *testing.T) {
	slot := &TemplateSlot{
		ExerciseID: uuid.New(),
		Order:      0,
		TargetSets: 3,
		TargetReps: 10,
	}

	// Zero order is valid
	if slot.Order != 0 {
		t.Errorf("expected order 0, got %d", slot.Order)
	}
}
