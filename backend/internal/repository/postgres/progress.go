package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
)

// ProgressRepository implements repository.ProgressRepository using PostgreSQL.
type ProgressRepository struct {
	db *sql.DB
}

// NewProgressRepository creates a new ProgressRepository.
func NewProgressRepository(db *sql.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// ExerciseHistory retrieves paginated session sets for a specific exercise, ordered by session start_at DESC.
func (r *ProgressRepository) ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, cursorStr string, limit int) ([]*domain.SessionSet, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("ws.user_id = $%d", argIdx))
	args = append(args, userID)
	argIdx++

	conditions = append(conditions, fmt.Sprintf("se.exercise_id = $%d", argIdx))
	args = append(args, exerciseID)
	argIdx++

	if cursorStr != "" {
		values, err := cursor.Decode(cursorStr, 2)
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
		conditions = append(conditions, fmt.Sprintf("(ws.start_at, ws.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursorTime, cursorID)
		argIdx += 2
	}

	whereClause := "WHERE " + joinConditions(conditions)

	query := fmt.Sprintf(`
		SELECT ss.id, ss.session_exercise_id, ss."order", ss.weight, ss.reps, ss.duration, ss.rpe,
		       ws.start_at, ws.id as session_id
		FROM session_sets ss
		JOIN session_exercises se ON se.id = ss.session_exercise_id
		JOIN workout_sessions ws ON ws.id = se.session_id
		%s
		ORDER BY ws.start_at DESC, ws.id DESC
		LIMIT $%d
	`, whereClause, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query exercise history: %w", err)
	}
	defer rows.Close()

	var sets []*domain.SessionSet
	for rows.Next() {
		set := &domain.SessionSet{}
		var startAt time.Time
		var sessionID uuid.UUID
		if err := rows.Scan(&set.ID, &set.SessionExerciseID, &set.Order, &set.Weight, &set.Reps, &set.Duration, &set.RPE, &startAt, &sessionID); err != nil {
			return nil, false, fmt.Errorf("scan set: %w", err)
		}
		sets = append(sets, set)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("rows iteration: %w", err)
	}

	hasMore := len(sets) > limit
	if hasMore {
		sets = sets[:limit]
	}

	return sets, hasMore, nil
}

// Summary retrieves aggregate statistics for a user's progress.
func (r *ProgressRepository) Summary(ctx context.Context, userID uuid.UUID) (*repository.ProgressSummary, error) {
	summary := &repository.ProgressSummary{}

	// Get total sessions and workouts
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT ws.id), COUNT(DISTINCT ws.template_id)
		FROM workout_sessions ws
		WHERE ws.user_id = $1
	`, userID).Scan(&summary.TotalSessions, &summary.TotalWorkouts)
	if err != nil {
		return nil, fmt.Errorf("query session counts: %w", err)
	}

	// Get total unique exercises
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT se.exercise_id)
		FROM session_exercises se
		JOIN workout_sessions ws ON ws.id = se.session_id
		WHERE ws.user_id = $1
	`, userID).Scan(&summary.TotalExercises)
	if err != nil {
		return nil, fmt.Errorf("query exercise count: %w", err)
	}

	// Get total time (sum of session durations)
	var totalTime sql.NullInt64
	err = r.db.QueryRowContext(ctx, `
		SELECT SUM(EXTRACT(EPOCH FROM (ws.end_at - ws.start_at)))
		FROM workout_sessions ws
		WHERE ws.user_id = $1 AND ws.end_at IS NOT NULL
	`, userID).Scan(&totalTime)
	if err != nil {
		return nil, fmt.Errorf("query total time: %w", err)
	}
	if totalTime.Valid {
		summary.TotalTime = int(totalTime.Int64)
	}

	// Get average session duration
	var avgDuration sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `
		SELECT AVG(EXTRACT(EPOCH FROM (ws.end_at - ws.start_at)))
		FROM workout_sessions ws
		WHERE ws.user_id = $1 AND ws.end_at IS NOT NULL
	`, userID).Scan(&avgDuration)
	if err != nil {
		return nil, fmt.Errorf("query avg duration: %w", err)
	}
	if avgDuration.Valid {
		summary.AvgSessionDuration = int(avgDuration.Float64)
	}

	return summary, nil
}
