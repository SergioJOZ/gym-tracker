package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
)

// jsonbMap is a helper type for scanning JSONB into a map[string]string.
// It also implements driver.Valuer so it can be passed directly to Exec/Query.
type jsonbMap map[string]string

// Scan implements the sql.Scanner interface for reading JSONB columns.
func (m *jsonbMap) Scan(src interface{}) error {
	if src == nil {
		*m = make(map[string]string)
		return nil
	}
	var source []byte
	switch v := src.(type) {
	case []byte:
		source = v
	case string:
		source = []byte(v)
	default:
		return fmt.Errorf("jsonbMap: unsupported type %T", src)
	}
	return json.Unmarshal(source, (*map[string]string)(m))
}

// Value implements driver.Valuer for writing JSONB columns.
func (m jsonbMap) Value() (interface{}, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

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
		INSERT INTO exercises (id, names, descriptions, muscle_group, equipment, difficulty, category, gif_path, thumbnail_path)
		VALUES ($1, $2::jsonb, $3::jsonb, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			names = EXCLUDED.names,
			descriptions = EXCLUDED.descriptions,
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
			ex.ID, jsonbMap(ex.NameByLang), jsonbMap(ex.DescriptionsByLang), ex.MuscleGroup,
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
		SELECT id, names, descriptions, muscle_group, equipment, difficulty, category, gif_path, thumbnail_path
		FROM exercises
		WHERE id = $1
	`

	ex := &domain.Exercise{}
	var names jsonbMap
	var descriptions jsonbMap
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ex.ID, &names, &descriptions, &ex.MuscleGroup,
		&ex.Equipment, &ex.Difficulty, &ex.Category,
		&ex.GIFPath, &ex.ThumbnailPath,
	)
	if err == nil {
		ex.NameByLang = map[string]string(names)
		ex.DescriptionsByLang = map[string]string(descriptions)
	}

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
		SELECT id, names, descriptions, muscle_group, equipment, difficulty, category, gif_path, thumbnail_path
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
		var names jsonbMap
		var descriptions jsonbMap
		if err := rows.Scan(
			&ex.ID, &names, &descriptions, &ex.MuscleGroup,
			&ex.Equipment, &ex.Difficulty, &ex.Category,
			&ex.GIFPath, &ex.ThumbnailPath,
		); err != nil {
			return nil, false, err
		}
		ex.NameByLang = map[string]string(names)
		ex.DescriptionsByLang = map[string]string(descriptions)
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

// Exists checks whether an exercise with the given ID exists.
func (r *ExerciseRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM exercises WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
