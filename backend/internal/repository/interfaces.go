package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// Create persists a new user to the database.
	Create(ctx context.Context, user *domain.User) error
	
	// FindByEmail retrieves a user by their email address.
	// Returns domain.ErrNotFound if the user doesn't exist.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	
	// FindByID retrieves a user by their ID.
	// Returns domain.ErrNotFound if the user doesn't exist.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// RefreshTokenRepository defines the interface for refresh token persistence operations.
type RefreshTokenRepository interface {
	// Create persists a new refresh token to the database.
	Create(ctx context.Context, token *domain.RefreshToken) error
	
	// FindByHash retrieves a refresh token by its hash.
	// Returns domain.ErrNotFound if the token doesn't exist.
	FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	
	// Revoke marks a refresh token as revoked by setting revoked_at.
	Revoke(ctx context.Context, id uuid.UUID) error
	
	// RevokeAllForUser revokes all refresh tokens for a given user.
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

// ExerciseFilter defines filtering criteria for listing exercises.
type ExerciseFilter struct {
	Search      string // full-text search query
	MuscleGroup string // filter by muscle group
	Equipment   string // filter by equipment type
	Difficulty  string // filter by difficulty level
	Category    string // filter by category
	Cursor      string // opaque cursor for pagination
	Limit       int    // page size
}

// ExerciseRepository defines the interface for exercise persistence operations.
type ExerciseRepository interface {
	// List retrieves exercises matching the filter with cursor-based pagination.
	// Returns a slice of exercises and whether more results exist.
	List(ctx context.Context, filter ExerciseFilter) ([]*domain.Exercise, bool, error)

	// GetByID retrieves an exercise by its ID.
	// Returns domain.ErrNotFound if the exercise doesn't exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error)

	// BulkUpsert inserts or updates a batch of exercises (for seeding).
	BulkUpsert(ctx context.Context, exercises []*domain.Exercise) error

	// Exists checks whether an exercise with the given ID exists.
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

// TemplateFilter defines filtering criteria for listing templates.
type TemplateFilter struct {
	Cursor string // opaque cursor for pagination (created_at DESC)
	Limit  int    // page size
}

// TemplateRepository defines the interface for workout template persistence operations.
type TemplateRepository interface {
	// Create persists a new workout template with its slots.
	Create(ctx context.Context, template *domain.WorkoutTemplate) error

	// Update replaces a template and its slots in a single transaction.
	Update(ctx context.Context, template *domain.WorkoutTemplate) error

	// Delete removes a template owned by the given user.
	// Returns domain.ErrNotFound if the template doesn't exist.
	Delete(ctx context.Context, userID, templateID uuid.UUID) error

	// FindByID retrieves a template with its slots, scoped to the given user.
	// Returns domain.ErrNotFound if the template doesn't exist.
	FindByID(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error)

	// List retrieves templates for a user with cursor-based pagination.
	// Returns a slice of templates (with slots), whether more results exist, and error.
	List(ctx context.Context, userID uuid.UUID, filter TemplateFilter) ([]*domain.WorkoutTemplate, bool, error)
}

// SessionFilter defines filtering criteria for listing workout sessions.
type SessionFilter struct {
	Cursor   string     // opaque cursor for pagination (start_at DESC)
	Limit    int        // page size
	StartDate *time.Time // optional: filter sessions starting on or after this date
	EndDate   *time.Time // optional: filter sessions starting before this date
}

// SessionRepository defines the interface for workout session persistence operations.
type SessionRepository interface {
	// Create persists a new workout session with its exercises and sets in a single transaction.
	Create(ctx context.Context, session *domain.WorkoutSession) error

	// Update replaces a session and its exercises/sets in a single transaction.
	Update(ctx context.Context, session *domain.WorkoutSession) error

	// Delete removes a session owned by the given user.
	// Returns domain.ErrNotFound if the session doesn't exist.
	Delete(ctx context.Context, userID, sessionID uuid.UUID) error

	// FindByID retrieves a session with its exercises and sets, scoped to the given user.
	// Returns domain.ErrNotFound if the session doesn't exist.
	FindByID(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error)

	// List retrieves sessions for a user with cursor-based pagination and optional date filtering.
	// Returns a slice of sessions (with nested data), whether more results exist, and error.
	List(ctx context.Context, userID uuid.UUID, filter SessionFilter) ([]*domain.WorkoutSession, bool, error)
}

// PersonalRecordRepository defines the interface for personal record persistence operations.
type PersonalRecordRepository interface {
	// Upsert inserts or updates a personal record using GREATEST to preserve the best values.
	Upsert(ctx context.Context, pr *domain.PersonalRecord) error

	// FindByUserAndExercise retrieves a personal record for a specific user and exercise.
	// Returns domain.ErrNotFound if no record exists.
	FindByUserAndExercise(ctx context.Context, userID, exerciseID uuid.UUID) (*domain.PersonalRecord, error)

	// FindByUser retrieves all personal records for a user.
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error)

	// RecalculateFromSessions recalculates personal records from all sessions for the given user and exercises.
	// This deletes existing PRs for those exercises and re-inserts from session data.
	RecalculateFromSessions(ctx context.Context, userID uuid.UUID, exerciseIDs []uuid.UUID) error
}
