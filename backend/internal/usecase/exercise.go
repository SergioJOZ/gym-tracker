package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
)

// ExerciseUseCase handles exercise catalog business logic.
type ExerciseUseCase struct {
	repo repository.ExerciseRepository
}

// NewExerciseUseCase creates a new ExerciseUseCase.
func NewExerciseUseCase(repo repository.ExerciseRepository) *ExerciseUseCase {
	return &ExerciseUseCase{repo: repo}
}

// List retrieves exercises matching the filter with cursor-based pagination.
func (uc *ExerciseUseCase) List(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
	// Apply default limit
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return uc.repo.List(ctx, filter)
}

// GetByID retrieves an exercise by its ID.
// Returns domain.ErrNotFound if the exercise doesn't exist.
func (uc *ExerciseUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
	return uc.repo.GetByID(ctx, id)
}
