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

// TemplateRepository implements repository.TemplateRepository using PostgreSQL.
type TemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new TemplateRepository.
func NewTemplateRepository(db *sql.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create persists a new workout template with its slots in a single transaction.
func (r *TemplateRepository) Create(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	tmpl.ID = uuid.New()
	tmpl.CreatedAt = now
	tmpl.UpdatedAt = now

	_, err = tx.ExecContext(ctx, `
		INSERT INTO workout_templates (id, user_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tmpl.ID, tmpl.UserID, tmpl.Name, tmpl.Description, tmpl.CreatedAt, tmpl.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert template: %w", err)
	}

	for i, slot := range tmpl.Slots {
		slot.ID = uuid.New()
		slot.TemplateID = tmpl.ID
		_, err = tx.ExecContext(ctx, `
			INSERT INTO template_slots (id, template_id, exercise_id, "order", target_sets, target_reps, target_weight, target_duration)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, slot.ID, slot.TemplateID, slot.ExerciseID, slot.Order, slot.TargetSets, slot.TargetReps, slot.TargetWeight, slot.TargetDuration)
		if err != nil {
			return fmt.Errorf("insert slot %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// Update replaces a template and its slots in a single transaction.
func (r *TemplateRepository) Update(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	tmpl.UpdatedAt = time.Now()

	result, err := tx.ExecContext(ctx, `
		UPDATE workout_templates
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND user_id = $5
	`, tmpl.Name, tmpl.Description, tmpl.UpdatedAt, tmpl.ID, tmpl.UserID)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}

	// Delete old slots and insert new ones
	_, err = tx.ExecContext(ctx, `DELETE FROM template_slots WHERE template_id = $1`, tmpl.ID)
	if err != nil {
		return fmt.Errorf("delete old slots: %w", err)
	}

	for i, slot := range tmpl.Slots {
		slot.ID = uuid.New()
		slot.TemplateID = tmpl.ID
		_, err = tx.ExecContext(ctx, `
			INSERT INTO template_slots (id, template_id, exercise_id, "order", target_sets, target_reps, target_weight, target_duration)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, slot.ID, slot.TemplateID, slot.ExerciseID, slot.Order, slot.TargetSets, slot.TargetReps, slot.TargetWeight, slot.TargetDuration)
		if err != nil {
			return fmt.Errorf("insert slot %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// Delete removes a template owned by the given user.
func (r *TemplateRepository) Delete(ctx context.Context, userID, templateID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM workout_templates WHERE id = $1 AND user_id = $2
	`, templateID, userID)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// FindByID retrieves a template with its slots, scoped to the given user.
func (r *TemplateRepository) FindByID(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error) {
	tmpl := &domain.WorkoutTemplate{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM workout_templates
		WHERE id = $1 AND user_id = $2
	`, templateID, userID).Scan(
		&tmpl.ID, &tmpl.UserID, &tmpl.Name, &tmpl.Description, &tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find template: %w", err)
	}

	// Load slots
	slots, err := r.loadSlots(ctx, tmpl.ID)
	if err != nil {
		return nil, err
	}
	tmpl.Slots = slots

	return tmpl, nil
}

// List retrieves templates for a user with cursor-based pagination (created_at DESC).
func (r *TemplateRepository) List(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
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
		// DESC ordering: we want rows where (created_at, id) < (cursorTime, cursorID)
		conditions = append(conditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursorTime, cursorID)
		argIdx += 2
	}

	whereClause := "WHERE " + joinConditions(conditions)

	query := fmt.Sprintf(`
		SELECT id, user_id, name, description, created_at, updated_at
		FROM workout_templates
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, whereClause, argIdx)
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var templates []*domain.WorkoutTemplate
	for rows.Next() {
		tmpl := &domain.WorkoutTemplate{}
		if err := rows.Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Name, &tmpl.Description, &tmpl.CreatedAt, &tmpl.UpdatedAt); err != nil {
			return nil, false, fmt.Errorf("scan template: %w", err)
		}
		templates = append(templates, tmpl)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("rows iteration: %w", err)
	}

	hasMore := len(templates) > limit
	if hasMore {
		templates = templates[:limit]
	}

	// Load slots for each template
	for _, tmpl := range templates {
		slots, err := r.loadSlots(ctx, tmpl.ID)
		if err != nil {
			return nil, false, err
		}
		tmpl.Slots = slots
	}

	return templates, hasMore, nil
}

// loadSlots retrieves all slots for a template, ordered by order ASC.
func (r *TemplateRepository) loadSlots(ctx context.Context, templateID uuid.UUID) ([]*domain.TemplateSlot, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, template_id, exercise_id, "order", target_sets, target_reps, target_weight, target_duration
		FROM template_slots
		WHERE template_id = $1
		ORDER BY "order" ASC, id ASC
	`, templateID)
	if err != nil {
		return nil, fmt.Errorf("load slots: %w", err)
	}
	defer rows.Close()

	var slots []*domain.TemplateSlot
	for rows.Next() {
		slot := &domain.TemplateSlot{}
		if err := rows.Scan(&slot.ID, &slot.TemplateID, &slot.ExerciseID, &slot.Order, &slot.TargetSets, &slot.TargetReps, &slot.TargetWeight, &slot.TargetDuration); err != nil {
			return nil, fmt.Errorf("scan slot: %w", err)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("slot rows iteration: %w", err)
	}

	return slots, nil
}

// joinConditions joins conditions with AND.
func joinConditions(conditions []string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}
