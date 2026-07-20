package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
)

// SessionUseCase handles workout session business logic.
type SessionUseCase struct {
	sessionRepo repository.SessionRepository
	prRepo      repository.PersonalRecordRepository
}

// NewSessionUseCase creates a new SessionUseCase.
func NewSessionUseCase(sessionRepo repository.SessionRepository, prRepo repository.PersonalRecordRepository) *SessionUseCase {
	return &SessionUseCase{
		sessionRepo: sessionRepo,
		prRepo:      prRepo,
	}
}

// Create validates a session, saves it, and calculates PRs inline.
func (uc *SessionUseCase) Create(ctx context.Context, session *domain.WorkoutSession) error {
	if err := session.Validate(); err != nil {
		return err
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return err
	}

	// Calculate and upsert PRs for each exercise in the session
	if uc.prRepo != nil {
		for _, ex := range session.Exercises {
			pr := uc.calculatePRFromExercise(session.UserID, ex)
			if pr != nil {
				if err := uc.prRepo.Upsert(ctx, pr); err != nil {
					return fmt.Errorf("upsert PR for exercise %s: %w", ex.ExerciseID, err)
				}
			}
		}
	}

	return nil
}

// Update validates a session, updates it, and recalculates PRs from all sessions for affected exercises.
func (uc *SessionUseCase) Update(ctx context.Context, session *domain.WorkoutSession) error {
	if err := session.Validate(); err != nil {
		return err
	}

	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return err
	}

	// Recalculate PRs from all sessions for affected exercises
	if uc.prRepo != nil {
		exerciseIDs := uc.extractExerciseIDs(session.Exercises)
		if len(exerciseIDs) > 0 {
			if err := uc.prRepo.RecalculateFromSessions(ctx, session.UserID, exerciseIDs); err != nil {
				return fmt.Errorf("recalculate PRs: %w", err)
			}
		}
	}

	return nil
}

// Delete verifies ownership, deletes the session, and recalculates PRs from remaining sessions.
func (uc *SessionUseCase) Delete(ctx context.Context, userID, sessionID uuid.UUID) error {
	// Get the session first to know which exercises are affected
	session, err := uc.sessionRepo.FindByID(ctx, userID, sessionID)
	if err != nil {
		return err
	}

	if err := uc.sessionRepo.Delete(ctx, userID, sessionID); err != nil {
		return err
	}

	// Recalculate PRs from remaining sessions for affected exercises
	if uc.prRepo != nil {
		exerciseIDs := uc.extractExerciseIDs(session.Exercises)
		if len(exerciseIDs) > 0 {
			if err := uc.prRepo.RecalculateFromSessions(ctx, userID, exerciseIDs); err != nil {
				return fmt.Errorf("recalculate PRs: %w", err)
			}
		}
	}

	return nil
}

// GetByID retrieves a session with nested data, scoped to the user.
func (uc *SessionUseCase) GetByID(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error) {
	return uc.sessionRepo.FindByID(ctx, userID, sessionID)
}

// List retrieves sessions for a user with cursor-based pagination.
func (uc *SessionUseCase) List(ctx context.Context, userID uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return uc.sessionRepo.List(ctx, userID, filter)
}

// calculatePRFromExercise computes personal record values from a session exercise.
func (uc *SessionUseCase) calculatePRFromExercise(userID uuid.UUID, ex *domain.SessionExercise) *domain.PersonalRecord {
	if len(ex.Sets) == 0 {
		return nil
	}

	pr := &domain.PersonalRecord{
		UserID:     userID,
		ExerciseID: ex.ExerciseID,
	}

	var maxWeight float64
	var maxReps int
	var maxVolume float64
	hasData := false

	for _, set := range ex.Sets {
		if set.Weight != nil && *set.Weight > maxWeight {
			maxWeight = *set.Weight
			pr.MaxWeight = set.Weight
			hasData = true
		}
		if set.Reps != nil && *set.Reps > maxReps {
			maxReps = *set.Reps
			pr.MaxReps = set.Reps
			hasData = true
		}
		if set.Weight != nil && set.Reps != nil {
			volume := *set.Weight * float64(*set.Reps)
			if volume > maxVolume {
				maxVolume = volume
				pr.MaxVolume = &volume
				hasData = true
			}
		}
	}

	if !hasData {
		return nil
	}

	return pr
}

// extractExerciseIDs extracts unique exercise IDs from a slice of session exercises.
func (uc *SessionUseCase) extractExerciseIDs(exercises []*domain.SessionExercise) []uuid.UUID {
	seen := make(map[uuid.UUID]bool)
	var ids []uuid.UUID
	for _, ex := range exercises {
		if !seen[ex.ExerciseID] {
			seen[ex.ExerciseID] = true
			ids = append(ids, ex.ExerciseID)
		}
	}
	return ids
}
