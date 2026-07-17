package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshTokenRepo_Create(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	// Create a user first
	userRepo := NewUserRepository(tdb.DB)
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	repo := NewRefreshTokenRepository(tdb.DB)

	token := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  "test-hash",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}

	err = repo.Create(context.Background(), token)
	require.NoError(t, err)
	assert.NotEmpty(t, token.ID)
}

func TestRefreshTokenRepo_FindByHash(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	// Create a user first
	userRepo := NewUserRepository(tdb.DB)
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	repo := NewRefreshTokenRepository(tdb.DB)

	token := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  "test-hash",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}

	err = repo.Create(context.Background(), token)
	require.NoError(t, err)

	found, err := repo.FindByHash(context.Background(), "test-hash")
	require.NoError(t, err)
	assert.Equal(t, token.ID, found.ID)
	assert.Equal(t, token.UserID, found.UserID)
	assert.Equal(t, token.TokenHash, found.TokenHash)
}

func TestRefreshTokenRepo_FindByHash_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewRefreshTokenRepository(tdb.DB)

	_, err := repo.FindByHash(context.Background(), "nonexistent-hash")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRefreshTokenRepo_Revoke(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	// Create a user first
	userRepo := NewUserRepository(tdb.DB)
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	repo := NewRefreshTokenRepository(tdb.DB)

	token := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  "test-hash",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}

	err = repo.Create(context.Background(), token)
	require.NoError(t, err)

	err = repo.Revoke(context.Background(), token.ID)
	require.NoError(t, err)

	found, err := repo.FindByHash(context.Background(), "test-hash")
	require.NoError(t, err)
	assert.NotNil(t, found.RevokedAt)
}

func TestRefreshTokenRepo_RevokeAllForUser(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	// Create a user first
	userRepo := NewUserRepository(tdb.DB)
	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	err := userRepo.Create(context.Background(), user)
	require.NoError(t, err)

	repo := NewRefreshTokenRepository(tdb.DB)

	// Create multiple tokens for the user
	token1 := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  "hash1",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}
	token2 := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		TokenHash:  "hash2",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}

	err = repo.Create(context.Background(), token1)
	require.NoError(t, err)
	err = repo.Create(context.Background(), token2)
	require.NoError(t, err)

	err = repo.RevokeAllForUser(context.Background(), user.ID)
	require.NoError(t, err)

	found1, err := repo.FindByHash(context.Background(), "hash1")
	require.NoError(t, err)
	assert.NotNil(t, found1.RevokedAt)

	found2, err := repo.FindByHash(context.Background(), "hash2")
	require.NoError(t, err)
	assert.NotNil(t, found2.RevokedAt)
}
