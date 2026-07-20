package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
)

// ExerciseRepository implements repository.ExerciseRepository using PostgreSQL.
type ExerciseRepository struct {
	db *sql.DB
}

// NewExerciseRepository creates a new ExerciseRepository.
func NewExerciseRepository(db *sql.DB) *ExerciseRepository {
	return &ExerciseRepository{db: db}
}

// BulkUpsert inserts or updates a batch of exercises using ON CONFLICT.
func (r *ExerciseRepository) BulkUpsert(ctx context.Context, exercises []*domain.Exercise) error {
	if len(exercises) == 0 {
		return nil
	}

	query := `
		INSERT INTO exercises (id, name, description, muscle_group, equipment, difficulty, category, gif_path, thumbnail_path)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			muscle_group = EXCLUDED.muscle_group,
			equipment = EXCLUDED.equipment,
			difficulty = EXCLUDED.difficulty,
			category = EXCLUDED.category,
			gif_path = EXCLUDED.gif_path,
			thumbnail_path = EXCLUDED.thumbnail_path,
			updated_at = NOW()
	`

	for _, ex := range exercises {
		_, err := r.db.ExecContext(ctx, query,
			ex.ID, ex.Name, ex.Description, ex.MuscleGroup,
			ex.Equipment, ex.Difficulty, ex.Category,
			ex.GIFPath, ex.ThumbnailPath,
		)
		if err != nil {
			return fmt.Errorf("bulk upsert exercise %s: %w", ex.ID, err)
		}
	}

	return nil
}

// GetByID retrieves an exercise by its ID.
func (r *ExerciseRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
	query := `
		SELECT id, name, description, muscle_group, equipment, difficulty, category, gif_path, thumbnail_path
		FROM exercises
		WHERE id = $1
	`

	ex := &domain.Exercise{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ex.ID, &ex.Name, &ex.Description, &ex.MuscleGroup,
		&ex.Equipment, &ex.Difficulty, &ex.Category,
		&ex.GIFPath, &ex.ThumbnailPath,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return ex, nil
}

// List retrieves exercises matching the filter with cursor-based pagination.
// It fetches limit+1 items to determine if more results exist.
func (r *ExerciseRepository) List(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Build dynamic query with filters
	var conditions []string
	var args []interface{}
	argIdx := 1

	// Full-text search
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("search_vector @@ plainto_tsquery('english', $%d)", argIdx))
		args = append(args, filter.Search)
		argIdx++
	}

	if filter.MuscleGroup != "" {
		conditions = append(conditions, fmt.Sprintf("muscle_group = $%d", argIdx))
		args = append(args, filter.MuscleGroup)
		argIdx++
	}

	if filter.Equipment != "" {
		conditions = append(conditions, fmt.Sprintf("equipment = $%d", argIdx))
		args = append(args, filter.Equipment)
		argIdx++
	}

	if filter.Difficulty != "" {
		conditions = append(conditions, fmt.Sprintf("difficulty = $%d", argIdx))
		args = append(args, filter.Difficulty)
		argIdx++
	}

	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}

	// Cursor pagination: keyset on (id)
	if filter.Cursor != "" {
		values, err := cursor.Decode(filter.Cursor, 1)
		if err != nil {
			return nil, false, fmt.Errorf("invalid cursor: %w", err)
		}
		cursorID, err := uuid.Parse(values[0])
		if err != nil {
			return nil, false, fmt.Errorf("invalid cursor ID: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf("id > $%d", argIdx))
		args = append(args, cursorID)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Fetch limit+1 to determine HasMore
	query := fmt.Sprintf(`
		SELECT id, name, description, muscle_group, equipment, difficulty, category, gif_path, thumbnail_path
		FROM exercises
		%s
		ORDER BY id ASC
		LIMIT $%d
	`, whereClause, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var exercises []*domain.Exercise
	for rows.Next() {
		ex := &domain.Exercise{}
		if err := rows.Scan(
			&ex.ID, &ex.Name, &ex.Description, &ex.MuscleGroup,
			&ex.Equipment, &ex.Difficulty, &ex.Category,
			&ex.GIFPath, &ex.ThumbnailPath,
		); err != nil {
			return nil, false, err
		}
		exercises = append(exercises, ex)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(exercises) > limit
	if hasMore {
		exercises = exercises[:limit]
	}

	return exercises, hasMore, nil
}
