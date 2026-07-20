package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
)

// TemplateUseCase handles workout template business logic.
type TemplateUseCase struct {
	templateRepo  repository.TemplateRepository
	exerciseRepo  repository.ExerciseRepository
}

// NewTemplateUseCase creates a new TemplateUseCase.
func NewTemplateUseCase(templateRepo repository.TemplateRepository, exerciseRepo repository.ExerciseRepository) *TemplateUseCase {
	return &TemplateUseCase{
		templateRepo: templateRepo,
		exerciseRepo: exerciseRepo,
	}
}

// Create validates a template, checks exercise existence, and persists it.
func (uc *TemplateUseCase) Create(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
	if err := tmpl.Validate(); err != nil {
		return err
	}

	if err := uc.validateExercises(ctx, tmpl.Slots); err != nil {
		return err
	}

	return uc.templateRepo.Create(ctx, tmpl)
}

// Update validates a template, checks exercise existence, and updates it.
func (uc *TemplateUseCase) Update(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
	if err := tmpl.Validate(); err != nil {
		return err
	}

	if err := uc.validateExercises(ctx, tmpl.Slots); err != nil {
		return err
	}

	return uc.templateRepo.Update(ctx, tmpl)
}

// Delete removes a template. The repository enforces user ownership.
func (uc *TemplateUseCase) Delete(ctx context.Context, userID, templateID uuid.UUID) error {
	return uc.templateRepo.Delete(ctx, userID, templateID)
}

// GetByID retrieves a template with its slots, scoped to the user.
func (uc *TemplateUseCase) GetByID(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error) {
	return uc.templateRepo.FindByID(ctx, userID, templateID)
}

// List retrieves templates for a user with cursor-based pagination.
func (uc *TemplateUseCase) List(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return uc.templateRepo.List(ctx, userID, filter)
}

// validateExercises checks that all exercise IDs in the slots exist.
func (uc *TemplateUseCase) validateExercises(ctx context.Context, slots []*domain.TemplateSlot) error {
	if uc.exerciseRepo == nil {
		return nil
	}

	for i, slot := range slots {
		exists, err := uc.exerciseRepo.Exists(ctx, slot.ExerciseID)
		if err != nil {
			return fmt.Errorf("check exercise %s: %w", slot.ExerciseID, err)
		}
		if !exists {
			return domain.NewAppError(
				"EXERCISE_NOT_FOUND",
				fmt.Sprintf("exercise at slot %d does not exist", i),
				400,
			)
		}
	}

	return nil
}
