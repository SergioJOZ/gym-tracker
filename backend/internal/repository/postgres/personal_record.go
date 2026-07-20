package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
)

// PersonalRecordRepository implements repository.PersonalRecordRepository using PostgreSQL.
type PersonalRecordRepository struct {
	db *sql.DB
}

// NewPersonalRecordRepository creates a new PersonalRecordRepository.
func NewPersonalRecordRepository(db *sql.DB) *PersonalRecordRepository {
	return &PersonalRecordRepository{db: db}
}

// Upsert inserts or updates a personal record using GREATEST to preserve the best values.
func (r *PersonalRecordRepository) Upsert(ctx context.Context, pr *domain.PersonalRecord) error {
	pr.ID = uuid.New()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO personal_records (id, user_id, exercise_id, max_weight, max_reps, max_volume, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, exercise_id) DO UPDATE SET
			max_weight = GREATEST(personal_records.max_weight, EXCLUDED.max_weight),
			max_reps = GREATEST(personal_records.max_reps, EXCLUDED.max_reps),
			max_volume = GREATEST(personal_records.max_volume, EXCLUDED.max_volume),
			updated_at = NOW()
		RETURNING id, updated_at
	`, pr.ID, pr.UserID, pr.ExerciseID, pr.MaxWeight, pr.MaxReps, pr.MaxVolume)
	if err != nil {
		return fmt.Errorf("upsert personal record: %w", err)
	}

	return nil
}

// FindByUserAndExercise retrieves a personal record for a specific user and exercise.
func (r *PersonalRecordRepository) FindByUserAndExercise(ctx context.Context, userID, exerciseID uuid.UUID) (*domain.PersonalRecord, error) {
	pr := &domain.PersonalRecord{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, exercise_id, max_weight, max_reps, max_volume, updated_at
		FROM personal_records
		WHERE user_id = $1 AND exercise_id = $2
	`, userID, exerciseID).Scan(
		&pr.ID, &pr.UserID, &pr.ExerciseID, &pr.MaxWeight, &pr.MaxReps, &pr.MaxVolume, &pr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find personal record: %w", err)
	}

	return pr, nil
}

// FindByUser retrieves all personal records for a user.
func (r *PersonalRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, exercise_id, max_weight, max_reps, max_volume, updated_at
		FROM personal_records
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("find personal records: %w", err)
	}
	defer rows.Close()

	var prs []*domain.PersonalRecord
	for rows.Next() {
		pr := &domain.PersonalRecord{}
		if err := rows.Scan(&pr.ID, &pr.UserID, &pr.ExerciseID, &pr.MaxWeight, &pr.MaxReps, &pr.MaxVolume, &pr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan personal record: %w", err)
		}
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return prs, nil
}

// RecalculateFromSessions recalculates personal records from all sessions for the given user and exercises.
func (r *PersonalRecordRepository) RecalculateFromSessions(ctx context.Context, userID uuid.UUID, exerciseIDs []uuid.UUID) error {
	if len(exerciseIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete existing PRs for these exercises
	for _, exerciseID := range exerciseIDs {
		_, err = tx.ExecContext(ctx, `
			DELETE FROM personal_records WHERE user_id = $1 AND exercise_id = $2
		`, userID, exerciseID)
		if err != nil {
			return fmt.Errorf("delete personal record: %w", err)
		}
	}

	// Recalculate from sessions
	for _, exerciseID := range exerciseIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO personal_records (id, user_id, exercise_id, max_weight, max_reps, max_volume, updated_at)
			SELECT
				gen_random_uuid(),
				$1,
				$2,
				MAX(ss.weight),
				MAX(ss.reps),
				MAX(COALESCE(ss.weight, 0) * COALESCE(ss.reps, 0)),
				NOW()
			FROM workout_sessions ws
			JOIN session_exercises se ON se.session_id = ws.id
			JOIN session_sets ss ON ss.session_exercise_id = se.id
			WHERE ws.user_id = $1 AND se.exercise_id = $2
				AND (ss.weight IS NOT NULL OR ss.reps IS NOT NULL)
			GROUP BY ws.user_id, se.exercise_id
		`, userID, exerciseID)
		if err != nil {
			return fmt.Errorf("recalculate personal record: %w", err)
		}
	}

	return tx.Commit()
}
