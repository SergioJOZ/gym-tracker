package repository

import (
	"context"

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
}
