package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
)

// SessionRepository implements repository.SessionRepository using PostgreSQL.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new SessionRepository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create persists a new workout session with its exercises and sets in a single transaction.
func (r *SessionRepository) Create(ctx context.Context, session *domain.WorkoutSession) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	session.ID = uuid.New()
	session.CreatedAt = now
	session.UpdatedAt = now

	_, err = tx.ExecContext(ctx, `
		INSERT INTO workout_sessions (id, user_id, template_id, name, notes, start_at, end_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, session.ID, session.UserID, session.TemplateID, session.Name, session.Notes, session.StartAt, session.EndAt, session.CreatedAt, session.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	for i, ex := range session.Exercises {
		ex.ID = uuid.New()
		ex.SessionID = session.ID
		_, err = tx.ExecContext(ctx, `
			INSERT INTO session_exercises (id, session_id, exercise_id, "order", notes)
			VALUES ($1, $2, $3, $4, $5)
		`, ex.ID, ex.SessionID, ex.ExerciseID, ex.Order, ex.Notes)
		if err != nil {
			return fmt.Errorf("insert exercise %d: %w", i, err)
		}

		for j, set := range ex.Sets {
			set.ID = uuid.New()
			set.SessionExerciseID = ex.ID
			_, err = tx.ExecContext(ctx, `
				INSERT INTO session_sets (id, session_exercise_id, "order", weight, reps, duration, rpe)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, set.ID, set.SessionExerciseID, set.Order, set.Weight, set.Reps, set.Duration, set.RPE)
			if err != nil {
				return fmt.Errorf("insert set %d for exercise %d: %w", j, i, err)
			}
		}
	}

	return tx.Commit()
}

// Update replaces a session and its exercises/sets in a single transaction.
func (r *SessionRepository) Update(ctx context.Context, session *domain.WorkoutSession) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	session.UpdatedAt = time.Now()

	result, err := tx.ExecContext(ctx, `
		UPDATE workout_sessions
		SET name = $1, notes = $2, start_at = $3, end_at = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7
	`, session.Name, session.Notes, session.StartAt, session.EndAt, session.UpdatedAt, session.ID, session.UserID)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}

	// Delete old exercises (cascade will delete sets)
	_, err = tx.ExecContext(ctx, `DELETE FROM session_exercises WHERE session_id = $1`, session.ID)
	if err != nil {
		return fmt.Errorf("delete old exercises: %w", err)
	}

	// Insert new exercises and sets
	for i, ex := range session.Exercises {
		ex.ID = uuid.New()
		ex.SessionID = session.ID
		_, err = tx.ExecContext(ctx, `
			INSERT INTO session_exercises (id, session_id, exercise_id, "order", notes)
			VALUES ($1, $2, $3, $4, $5)
		`, ex.ID, ex.SessionID, ex.ExerciseID, ex.Order, ex.Notes)
		if err != nil {
			return fmt.Errorf("insert exercise %d: %w", i, err)
		}

		for j, set := range ex.Sets {
			set.ID = uuid.New()
			set.SessionExerciseID = ex.ID
			_, err = tx.ExecContext(ctx, `
				INSERT INTO session_sets (id, session_exercise_id, "order", weight, reps, duration, rpe)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, set.ID, set.SessionExerciseID, set.Order, set.Weight, set.Reps, set.Duration, set.RPE)
			if err != nil {
				return fmt.Errorf("insert set %d for exercise %d: %w", j, i, err)
			}
		}
	}

	return tx.Commit()
}

// Delete removes a session owned by the given user.
func (r *SessionRepository) Delete(ctx context.Context, userID, sessionID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM workout_sessions WHERE id = $1 AND user_id = $2
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// FindByID retrieves a session with its exercises and sets, scoped to the given user.
func (r *SessionRepository) FindByID(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error) {
	session := &domain.WorkoutSession{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, template_id, name, notes, start_at, end_at, created_at, updated_at
		FROM workout_sessions
		WHERE id = $1 AND user_id = $2
	`, sessionID, userID).Scan(
		&session.ID, &session.UserID, &session.TemplateID, &session.Name, &session.Notes,
		&session.StartAt, &session.EndAt, &session.CreatedAt, &session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find session: %w", err)
	}

	// Load exercises
	exercises, err := r.loadExercises(ctx, session.ID)
	if err != nil {
		return nil, err
	}
	session.Exercises = exercises

	return session, nil
}

// List retrieves sessions for a user with cursor-based pagination and optional date filtering.
func (r *SessionRepository) List(ctx context.Context, userID uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
	args = append(args, userID)
	argIdx++

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("start_at >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("start_at < $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if filter.Cursor != "" {
		values, err := cursor.Decode(filter.Cursor, 2)
		if err != nil {
			return nil, false, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorTime, err := time.Parse(time.RFC3339Nano, values[0])
		if err != nil {
			return nil, false, fmt.Errorf("invalid cursor time: %w", err)
		}
		cursorID, err := uuid.Parse(values[1])
		if err != nil {
			return nil, false, fmt.Errorf("invalid cursor ID: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf("(start_at, id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursorTime, cursorID)
		argIdx += 2
	}

	whereClause := "WHERE " + joinConditions(conditions)

	query := fmt.Sprintf(`
		SELECT id, user_id, template_id, name, notes, start_at, end_at, created_at, updated_at
		FROM workout_sessions
		%s
		ORDER BY start_at DESC, id DESC
		LIMIT $%d
	`, whereClause, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.WorkoutSession
	for rows.Next() {
		session := &domain.WorkoutSession{}
		if err := rows.Scan(
			&session.ID, &session.UserID, &session.TemplateID, &session.Name, &session.Notes,
			&session.StartAt, &session.EndAt, &session.CreatedAt, &session.UpdatedAt,
		); err != nil {
			return nil, false, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("rows iteration: %w", err)
	}

	hasMore := len(sessions) > limit
	if hasMore {
		sessions = sessions[:limit]
	}

	// Load exercises and sets for each session
	for _, session := range sessions {
		exercises, err := r.loadExercises(ctx, session.ID)
		if err != nil {
			return nil, false, err
		}
		session.Exercises = exercises
	}

	return sessions, hasMore, nil
}

// loadExercises retrieves all exercises for a session with their sets.
func (r *SessionRepository) loadExercises(ctx context.Context, sessionID uuid.UUID) ([]*domain.SessionExercise, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, session_id, exercise_id, "order", notes
		FROM session_exercises
		WHERE session_id = $1
		ORDER BY "order" ASC, id ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("load exercises: %w", err)
	}
	defer rows.Close()

	var exercises []*domain.SessionExercise
	for rows.Next() {
		ex := &domain.SessionExercise{}
		if err := rows.Scan(&ex.ID, &ex.SessionID, &ex.ExerciseID, &ex.Order, &ex.Notes); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}

		// Load sets for this exercise
		sets, err := r.loadSets(ctx, ex.ID)
		if err != nil {
			return nil, err
		}
		ex.Sets = sets

		exercises = append(exercises, ex)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("exercise rows iteration: %w", err)
	}

	return exercises, nil
}

// loadSets retrieves all sets for a session exercise.
func (r *SessionRepository) loadSets(ctx context.Context, exerciseID uuid.UUID) ([]*domain.SessionSet, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, session_exercise_id, "order", weight, reps, duration, rpe
		FROM session_sets
		WHERE session_exercise_id = $1
		ORDER BY "order" ASC, id ASC
	`, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("load sets: %w", err)
	}
	defer rows.Close()

	var sets []*domain.SessionSet
	for rows.Next() {
		set := &domain.SessionSet{}
		if err := rows.Scan(&set.ID, &set.SessionExerciseID, &set.Order, &set.Weight, &set.Reps, &set.Duration, &set.RPE); err != nil {
			return nil, fmt.Errorf("scan set: %w", err)
		}
		sets = append(sets, set)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("set rows iteration: %w", err)
	}

	return sets, nil
}
