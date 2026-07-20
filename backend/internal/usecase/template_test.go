package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTemplateRepository is a mock implementation of TemplateRepository for testing.
type MockTemplateRepository struct {
	CreateFunc  func(ctx context.Context, t *domain.WorkoutTemplate) error
	UpdateFunc  func(ctx context.Context, t *domain.WorkoutTemplate) error
	DeleteFunc  func(ctx context.Context, userID, templateID uuid.UUID) error
	FindByIDFunc func(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error)
	ListFunc    func(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error)
}

func (m *MockTemplateRepository) Create(ctx context.Context, t *domain.WorkoutTemplate) error {
	return m.CreateFunc(ctx, t)
}
func (m *MockTemplateRepository) Update(ctx context.Context, t *domain.WorkoutTemplate) error {
	return m.UpdateFunc(ctx, t)
}
func (m *MockTemplateRepository) Delete(ctx context.Context, userID, templateID uuid.UUID) error {
	return m.DeleteFunc(ctx, userID, templateID)
}
func (m *MockTemplateRepository) FindByID(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error) {
	return m.FindByIDFunc(ctx, userID, templateID)
}
func (m *MockTemplateRepository) List(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
	return m.ListFunc(ctx, userID, filter)
}

func TestTemplateUseCase_Create_Success(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockTemplateRepo := &MockTemplateRepository{
		CreateFunc: func(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
			assert.Equal(t, userID, tmpl.UserID)
			assert.Equal(t, "Push Day", tmpl.Name)
			assert.Len(t, tmpl.Slots, 1)
			tmpl.ID = uuid.New()
			return nil
		},
	}

	mockExerciseRepo := &MockExerciseRepository{
		ExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			assert.Equal(t, exerciseID, id)
			return true, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, mockExerciseRepo)

	tmpl := &domain.WorkoutTemplate{
		UserID: userID,
		Name:   "Push Day",
		Slots: []*domain.TemplateSlot{
			{ExerciseID: exerciseID, Order: 0, TargetSets: 3, TargetReps: 10},
		},
	}

	err := uc.Create(context.Background(), tmpl)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tmpl.ID)
}

func TestTemplateUseCase_Create_ValidationFails(t *testing.T) {
	mockTemplateRepo := &MockTemplateRepository{}
	mockExerciseRepo := &MockExerciseRepository{}

	uc := NewTemplateUseCase(mockTemplateRepo, mockExerciseRepo)

	// Empty name should fail validation
	tmpl := &domain.WorkoutTemplate{
		UserID: uuid.New(),
		Name:   "",
	}

	err := uc.Create(context.Background(), tmpl)
	require.Error(t, err)

	appErr, ok := err.(*domain.AppError)
	require.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestTemplateUseCase_Create_ExerciseNotFound(t *testing.T) {
	userID := uuid.New()
	nonExistentExercise := uuid.New()

	mockTemplateRepo := &MockTemplateRepository{}
	mockExerciseRepo := &MockExerciseRepository{
		ExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, mockExerciseRepo)

	tmpl := &domain.WorkoutTemplate{
		UserID: userID,
		Name:   "Push Day",
		Slots: []*domain.TemplateSlot{
			{ExerciseID: nonExistentExercise, Order: 0},
		},
	}

	err := uc.Create(context.Background(), tmpl)
	require.Error(t, err)

	appErr, ok := err.(*domain.AppError)
	require.True(t, ok)
	assert.Equal(t, "EXERCISE_NOT_FOUND", appErr.Code)
}

func TestTemplateUseCase_Create_TooManySlots(t *testing.T) {
	userID := uuid.New()
	slots := make([]*domain.TemplateSlot, 51)
	for i := range slots {
		slots[i] = &domain.TemplateSlot{ExerciseID: uuid.New(), Order: i}
	}

	mockTemplateRepo := &MockTemplateRepository{}
	mockExerciseRepo := &MockExerciseRepository{
		ExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, mockExerciseRepo)

	tmpl := &domain.WorkoutTemplate{
		UserID: userID,
		Name:   "Mega Template",
		Slots:  slots,
	}

	err := uc.Create(context.Background(), tmpl)
	require.Error(t, err)

	appErr, ok := err.(*domain.AppError)
	require.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestTemplateUseCase_Update_Success(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()
	exerciseID := uuid.New()

	mockTemplateRepo := &MockTemplateRepository{
		UpdateFunc: func(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
			assert.Equal(t, templateID, tmpl.ID)
			assert.Equal(t, userID, tmpl.UserID)
			assert.Equal(t, "Updated", tmpl.Name)
			return nil
		},
	}

	mockExerciseRepo := &MockExerciseRepository{
		ExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, mockExerciseRepo)

	tmpl := &domain.WorkoutTemplate{
		ID:     templateID,
		UserID: userID,
		Name:   "Updated",
		Slots: []*domain.TemplateSlot{
			{ExerciseID: exerciseID, Order: 0},
		},
	}

	err := uc.Update(context.Background(), tmpl)
	require.NoError(t, err)
}

func TestTemplateUseCase_Update_NotFound(t *testing.T) {
	userID := uuid.New()

	mockTemplateRepo := &MockTemplateRepository{
		UpdateFunc: func(ctx context.Context, tmpl *domain.WorkoutTemplate) error {
			return domain.ErrNotFound
		},
	}

	mockExerciseRepo := &MockExerciseRepository{
		ExistsFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, mockExerciseRepo)

	tmpl := &domain.WorkoutTemplate{
		ID:     uuid.New(),
		UserID: userID,
		Name:   "Ghost",
	}

	err := uc.Update(context.Background(), tmpl)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTemplateUseCase_Delete_Success(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()

	mockTemplateRepo := &MockTemplateRepository{
		DeleteFunc: func(ctx context.Context, uid, tid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, templateID, tid)
			return nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, nil)

	err := uc.Delete(context.Background(), userID, templateID)
	require.NoError(t, err)
}

func TestTemplateUseCase_Delete_NotFound(t *testing.T) {
	mockTemplateRepo := &MockTemplateRepository{
		DeleteFunc: func(ctx context.Context, uid, tid uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, nil)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestTemplateUseCase_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	templateID := uuid.New()
	expected := &domain.WorkoutTemplate{
		ID:     templateID,
		UserID: userID,
		Name:   "My Template",
	}

	mockTemplateRepo := &MockTemplateRepository{
		FindByIDFunc: func(ctx context.Context, uid, tid uuid.UUID) (*domain.WorkoutTemplate, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, templateID, tid)
			return expected, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, nil)

	result, err := uc.GetByID(context.Background(), userID, templateID)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestTemplateUseCase_GetByID_NotFound(t *testing.T) {
	mockTemplateRepo := &MockTemplateRepository{
		FindByIDFunc: func(ctx context.Context, uid, tid uuid.UUID) (*domain.WorkoutTemplate, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, nil)

	result, err := uc.GetByID(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)
}

func TestTemplateUseCase_List_Success(t *testing.T) {
	userID := uuid.New()
	templates := []*domain.WorkoutTemplate{
		{ID: uuid.New(), UserID: userID, Name: "T1"},
		{ID: uuid.New(), UserID: userID, Name: "T2"},
	}

	mockTemplateRepo := &MockTemplateRepository{
		ListFunc: func(ctx context.Context, uid uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, 10, filter.Limit)
			return templates, false, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, nil)

	result, hasMore, err := uc.List(context.Background(), userID, repository.TemplateFilter{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.False(t, hasMore)
}

func TestTemplateUseCase_List_DefaultLimit(t *testing.T) {
	mockTemplateRepo := &MockTemplateRepository{
		ListFunc: func(ctx context.Context, uid uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error) {
			assert.Equal(t, 20, filter.Limit)
			return []*domain.WorkoutTemplate{}, false, nil
		},
	}

	uc := NewTemplateUseCase(mockTemplateRepo, nil)

	_, _, err := uc.List(context.Background(), uuid.New(), repository.TemplateFilter{})
	require.NoError(t, err)
}
