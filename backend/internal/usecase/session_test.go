package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSessionRepository is a mock implementation of SessionRepository for testing.
type MockSessionRepository struct {
	CreateFunc  func(ctx context.Context, session *domain.WorkoutSession) error
	UpdateFunc  func(ctx context.Context, session *domain.WorkoutSession) error
	DeleteFunc  func(ctx context.Context, userID, sessionID uuid.UUID) error
	FindByIDFunc func(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error)
	ListFunc    func(ctx context.Context, userID uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error)
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.WorkoutSession) error {
	return m.CreateFunc(ctx, session)
}
func (m *MockSessionRepository) Update(ctx context.Context, session *domain.WorkoutSession) error {
	return m.UpdateFunc(ctx, session)
}
func (m *MockSessionRepository) Delete(ctx context.Context, userID, sessionID uuid.UUID) error {
	return m.DeleteFunc(ctx, userID, sessionID)
}
func (m *MockSessionRepository) FindByID(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error) {
	return m.FindByIDFunc(ctx, userID, sessionID)
}
func (m *MockSessionRepository) List(ctx context.Context, userID uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error) {
	return m.ListFunc(ctx, userID, filter)
}

// MockPersonalRecordRepository is a mock implementation of PersonalRecordRepository for testing.
type MockPersonalRecordRepository struct {
	UpsertFunc                    func(ctx context.Context, pr *domain.PersonalRecord) error
	FindByUserAndExerciseFunc     func(ctx context.Context, userID, exerciseID uuid.UUID) (*domain.PersonalRecord, error)
	FindByUserFunc                func(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error)
	RecalculateFromSessionsFunc   func(ctx context.Context, userID uuid.UUID, exerciseIDs []uuid.UUID) error
}

func (m *MockPersonalRecordRepository) Upsert(ctx context.Context, pr *domain.PersonalRecord) error {
	return m.UpsertFunc(ctx, pr)
}
func (m *MockPersonalRecordRepository) FindByUserAndExercise(ctx context.Context, userID, exerciseID uuid.UUID) (*domain.PersonalRecord, error) {
	return m.FindByUserAndExerciseFunc(ctx, userID, exerciseID)
}
func (m *MockPersonalRecordRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error) {
	return m.FindByUserFunc(ctx, userID)
}
func (m *MockPersonalRecordRepository) RecalculateFromSessions(ctx context.Context, userID uuid.UUID, exerciseIDs []uuid.UUID) error {
	return m.RecalculateFromSessionsFunc(ctx, userID, exerciseIDs)
}

func TestSessionUseCase_Create_Success(t *testing.T) {
	userID := uuid.New()
	exerciseID := uuid.New()

	mockSessionRepo := &MockSessionRepository{
		CreateFunc: func(ctx context.Context, session *domain.WorkoutSession) error {
			assert.Equal(t, userID, session.UserID)
			assert.Equal(t, "Morning Push", session.Name)
			session.ID = uuid.New()
			return nil
		},
	}

	mockPRRepo := &MockPersonalRecordRepository{
		UpsertFunc: func(ctx context.Context, pr *domain.PersonalRecord) error {
			assert.Equal(t, userID, pr.UserID)
			assert.Equal(t, exerciseID, pr.ExerciseID)
			return nil
		},
	}

	uc := NewSessionUseCase(mockSessionRepo, mockPRRepo)

	session := &domain.WorkoutSession{
		UserID:  userID,
		Name:    "Morning Push",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(80.0), Reps: intPtr(10)},
				},
			},
		},
	}

	err := uc.Create(context.Background(), session)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
}

func TestSessionUseCase_Create_ValidationFails(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockPRRepo := &MockPersonalRecordRepository{}

	uc := NewSessionUseCase(mockSessionRepo, mockPRRepo)

	// Empty name should fail validation
	session := &domain.WorkoutSession{
		UserID: uuid.New(),
		Name:   "",
	}

	err := uc.Create(context.Background(), session)
	require.Error(t, err)

	appErr, ok := err.(*domain.AppError)
	require.True(t, ok)
	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
}

func TestSessionUseCase_Update_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	exerciseID := uuid.New()

	mockSessionRepo := &MockSessionRepository{
		UpdateFunc: func(ctx context.Context, session *domain.WorkoutSession) error {
			assert.Equal(t, sessionID, session.ID)
			assert.Equal(t, userID, session.UserID)
			return nil
		},
	}

	mockPRRepo := &MockPersonalRecordRepository{
		RecalculateFromSessionsFunc: func(ctx context.Context, uid uuid.UUID, exerciseIDs []uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Contains(t, exerciseIDs, exerciseID)
			return nil
		},
	}

	uc := NewSessionUseCase(mockSessionRepo, mockPRRepo)

	session := &domain.WorkoutSession{
		ID:      sessionID,
		UserID:  userID,
		Name:    "Updated Session",
		StartAt: time.Now(),
		Exercises: []*domain.SessionExercise{
			{
				ExerciseID: exerciseID,
				Order:      0,
				Sets: []*domain.SessionSet{
					{Order: 0, Weight: floatPtr(90.0), Reps: intPtr(8)},
				},
			},
		},
	}

	err := uc.Update(context.Background(), session)
	require.NoError(t, err)
}

func TestSessionUseCase_Delete_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	exerciseID := uuid.New()

	mockSessionRepo := &MockSessionRepository{
		FindByIDFunc: func(ctx context.Context, uid, sid uuid.UUID) (*domain.WorkoutSession, error) {
			return &domain.WorkoutSession{
				ID:     sessionID,
				UserID: userID,
				Exercises: []*domain.SessionExercise{
					{ExerciseID: exerciseID},
				},
			}, nil
		},
		DeleteFunc: func(ctx context.Context, uid, sid uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, sessionID, sid)
			return nil
		},
	}

	mockPRRepo := &MockPersonalRecordRepository{
		RecalculateFromSessionsFunc: func(ctx context.Context, uid uuid.UUID, exerciseIDs []uuid.UUID) error {
			assert.Equal(t, userID, uid)
			assert.Contains(t, exerciseIDs, exerciseID)
			return nil
		},
	}

	uc := NewSessionUseCase(mockSessionRepo, mockPRRepo)

	err := uc.Delete(context.Background(), userID, sessionID)
	require.NoError(t, err)
}

func TestSessionUseCase_GetByID_Success(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	expected := &domain.WorkoutSession{
		ID:     sessionID,
		UserID: userID,
		Name:   "My Session",
	}

	mockSessionRepo := &MockSessionRepository{
		FindByIDFunc: func(ctx context.Context, uid, sid uuid.UUID) (*domain.WorkoutSession, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, sessionID, sid)
			return expected, nil
		},
	}

	uc := NewSessionUseCase(mockSessionRepo, nil)

	result, err := uc.GetByID(context.Background(), userID, sessionID)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestSessionUseCase_List_Success(t *testing.T) {
	userID := uuid.New()
	sessions := []*domain.WorkoutSession{
		{ID: uuid.New(), UserID: userID, Name: "S1"},
		{ID: uuid.New(), UserID: userID, Name: "S2"},
	}

	mockSessionRepo := &MockSessionRepository{
		ListFunc: func(ctx context.Context, uid uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, 10, filter.Limit)
			return sessions, false, nil
		},
	}

	uc := NewSessionUseCase(mockSessionRepo, nil)

	result, hasMore, err := uc.List(context.Background(), userID, repository.SessionFilter{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.False(t, hasMore)
}

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int          { return &i }
