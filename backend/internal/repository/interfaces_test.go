package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
)

// mockTemplateRepo verifies the TemplateRepository interface is usable.
type mockTemplateRepo struct{}

func (m *mockTemplateRepo) Create(ctx context.Context, t *domain.WorkoutTemplate) error {
	return nil
}
func (m *mockTemplateRepo) Update(ctx context.Context, t *domain.WorkoutTemplate) error {
	return nil
}
func (m *mockTemplateRepo) Delete(ctx context.Context, userID, templateID uuid.UUID) error {
	return nil
}
func (m *mockTemplateRepo) FindByID(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error) {
	return nil, nil
}
func (m *mockTemplateRepo) List(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
	return nil, false, nil
}

// Verify mockTemplateRepo satisfies the interface at compile time.
var _ repository.TemplateRepository = (*mockTemplateRepo)(nil)

func TestTemplateRepository_InterfaceCompiles(t *testing.T) {
	var repo repository.TemplateRepository = &mockTemplateRepo{}
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}
}

func TestTemplateFilter_Defaults(t *testing.T) {
	f := repository.TemplateFilter{}
	if f.Limit != 0 {
		t.Errorf("expected default limit 0, got %d", f.Limit)
	}
	if f.Cursor != "" {
		t.Errorf("expected empty cursor, got %q", f.Cursor)
	}
}

// mockExerciseRepoForExists verifies the Exists method on ExerciseRepository.
type mockExerciseRepoForExists struct {
	existsFunc func(ctx context.Context, id uuid.UUID) (bool, error)
}

func (m *mockExerciseRepoForExists) List(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error) {
	return nil, false, nil
}
func (m *mockExerciseRepoForExists) GetByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error) {
	return nil, nil
}
func (m *mockExerciseRepoForExists) BulkUpsert(ctx context.Context, exercises []*domain.Exercise) error {
	return nil
}
func (m *mockExerciseRepoForExists) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return m.existsFunc(ctx, id)
}

var _ repository.ExerciseRepository = (*mockExerciseRepoForExists)(nil)

func TestExerciseRepository_Exists_Interface(t *testing.T) {
	var repo repository.ExerciseRepository = &mockExerciseRepoForExists{
		existsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
	}

	exists, err := repo.Exists(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists to be true")
	}
}
