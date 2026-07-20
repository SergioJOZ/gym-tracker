package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
)

// ProgressUseCase handles progress and history business logic.
type ProgressUseCase struct {
	progressRepo repository.ProgressRepository
	prRepo       repository.PersonalRecordRepository
}

// NewProgressUseCase creates a new ProgressUseCase.
func NewProgressUseCase(progressRepo repository.ProgressRepository, prRepo repository.PersonalRecordRepository) *ProgressUseCase {
	return &ProgressUseCase{
		progressRepo: progressRepo,
		prRepo:       prRepo,
	}
}

// ListPRs retrieves all personal records for a user.
func (uc *ProgressUseCase) ListPRs(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error) {
	return uc.prRepo.FindByUser(ctx, userID)
}

// ExerciseHistory retrieves paginated session sets for a specific exercise.
func (uc *ProgressUseCase) ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return uc.progressRepo.ExerciseHistory(ctx, userID, exerciseID, cursor, limit)
}

// Summary retrieves aggregate statistics for a user's progress.
func (uc *ProgressUseCase) Summary(ctx context.Context, userID uuid.UUID) (*repository.ProgressSummary, error) {
	return uc.progressRepo.Summary(ctx, userID)
}
