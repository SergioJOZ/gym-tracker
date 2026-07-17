package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	jwtPkg "github.com/sergiojoz/gym-tracker/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	CreateFunc      func(ctx context.Context, user *domain.User) error
	FindByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	FindByIDFunc    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.CreateFunc(ctx, user)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.FindByEmailFunc(ctx, email)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return m.FindByIDFunc(ctx, id)
}

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository for testing.
type MockRefreshTokenRepository struct {
	CreateFunc         func(ctx context.Context, token *domain.RefreshToken) error
	FindByHashFunc     func(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RevokeFunc         func(ctx context.Context, id uuid.UUID) error
	RevokeAllForUserFunc func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	return m.CreateFunc(ctx, token)
}

func (m *MockRefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	return m.FindByHashFunc(ctx, hash)
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	return m.RevokeFunc(ctx, id)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return m.RevokeAllForUserFunc(ctx, userID)
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	userID := uuid.New()
	mockUserRepo := &MockUserRepository{
		CreateFunc: func(ctx context.Context, user *domain.User) error {
			user.ID = userID
			return nil
		},
		FindByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := NewAuthUseCase(mockUserRepo, nil, nil)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := uc.Register(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, userID, resp.User.ID)
	assert.Equal(t, "test@example.com", resp.User.Email)
}

func TestAuthUseCase_Register_DuplicateEmail(t *testing.T) {
	existingUser := &domain.User{
		ID:    uuid.New(),
		Email: "existing@example.com",
	}

	mockUserRepo := &MockUserRepository{
		CreateFunc: func(ctx context.Context, user *domain.User) error {
			return domain.ErrConflict
		},
		FindByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return existingUser, nil
		},
	}

	uc := NewAuthUseCase(mockUserRepo, nil, nil)

	req := &RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	_, err := uc.Register(context.Background(), req)
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestAuthUseCase_Register_InvalidEmail(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	uc := NewAuthUseCase(mockUserRepo, nil, nil)

	req := &RegisterRequest{
		Email:    "not-an-email",
		Password: "password123",
	}

	_, err := uc.Register(context.Background(), req)
	assert.Error(t, err)
}

func TestAuthUseCase_Register_PasswordTooShort(t *testing.T) {
	mockUserRepo := &MockUserRepository{}
	uc := NewAuthUseCase(mockUserRepo, nil, nil)

	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "short",
	}

	_, err := uc.Register(context.Background(), req)
	assert.Error(t, err)
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	userID := uuid.New()
	hashedPassword := "$2a$10$xJTYA.Jd2Zzw1GvfqWY03OAZIOZwz1RuXRPsSyTRu5HF.bds5Og2G" // bcrypt hash of "password123"

	existingUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: hashedPassword,
	}

	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return existingUser, nil
		},
	}

	mockTokenRepo := &MockRefreshTokenRepository{
		CreateFunc: func(ctx context.Context, token *domain.RefreshToken) error {
			return nil
		},
	}

	jwtCfg := &jwtPkg.Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(mockUserRepo, mockTokenRepo, jwtCfg)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := uc.Login(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestAuthUseCase_Login_WrongPassword(t *testing.T) {
	userID := uuid.New()
	hashedPassword := "$2a$10$xJTYA.Jd2Zzw1GvfqWY03OAZIOZwz1RuXRPsSyTRu5HF.bds5Og2G" // bcrypt hash of "password123"

	existingUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: hashedPassword,
	}

	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return existingUser, nil
		},
	}

	jwtCfg := &jwtPkg.Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(mockUserRepo, nil, jwtCfg)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err := uc.Login(context.Background(), req)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestAuthUseCase_Login_UserNotFound(t *testing.T) {
	mockUserRepo := &MockUserRepository{
		FindByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}

	jwtCfg := &jwtPkg.Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(mockUserRepo, nil, jwtCfg)

	req := &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := uc.Login(context.Background(), req)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestAuthUseCase_Refresh_Success(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()

	existingToken := &domain.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: "old-hash",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mockTokenRepo := &MockRefreshTokenRepository{
		FindByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return existingToken, nil
		},
		RevokeFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
		CreateFunc: func(ctx context.Context, token *domain.RefreshToken) error {
			return nil
		},
	}

	jwtCfg := &jwtPkg.Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(nil, mockTokenRepo, jwtCfg)

	req := &RefreshRequest{
		RefreshToken: "valid-refresh-token",
	}

	resp, err := uc.Refresh(context.Background(), req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func TestAuthUseCase_Refresh_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()

	existingToken := &domain.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: "old-hash",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}

	mockTokenRepo := &MockRefreshTokenRepository{
		FindByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return existingToken, nil
		},
	}

	jwtCfg := &jwtPkg.Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(nil, mockTokenRepo, jwtCfg)

	req := &RefreshRequest{
		RefreshToken: "expired-token",
	}

	_, err := uc.Refresh(context.Background(), req)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestAuthUseCase_Refresh_RevokedToken(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()
	revokedAt := time.Now()

	existingToken := &domain.RefreshToken{
		ID:         tokenID,
		UserID:     userID,
		TokenHash:  "old-hash",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		RevokedAt:  &revokedAt,
	}

	mockTokenRepo := &MockRefreshTokenRepository{
		FindByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return existingToken, nil
		},
	}

	jwtCfg := &jwtPkg.Config{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	uc := NewAuthUseCase(nil, mockTokenRepo, jwtCfg)

	req := &RefreshRequest{
		RefreshToken: "revoked-token",
	}

	_, err := uc.Refresh(context.Background(), req)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestAuthUseCase_Logout_Success(t *testing.T) {
	userID := uuid.New()
	tokenID := uuid.New()

	existingToken := &domain.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: "hash",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mockTokenRepo := &MockRefreshTokenRepository{
		FindByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return existingToken, nil
		},
		RevokeFunc: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	uc := NewAuthUseCase(nil, mockTokenRepo, nil)

	req := &LogoutRequest{
		RefreshToken: "valid-token",
	}

	err := uc.Logout(context.Background(), req)
	require.NoError(t, err)
}

func TestAuthUseCase_Logout_TokenNotFound(t *testing.T) {
	mockTokenRepo := &MockRefreshTokenRepository{
		FindByHashFunc: func(ctx context.Context, hash string) (*domain.RefreshToken, error) {
			return nil, domain.ErrNotFound
		},
	}

	uc := NewAuthUseCase(nil, mockTokenRepo, nil)

	req := &LogoutRequest{
		RefreshToken: "nonexistent-token",
	}

	err := uc.Logout(context.Background(), req)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}
