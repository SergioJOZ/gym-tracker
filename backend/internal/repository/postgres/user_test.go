package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_Create(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.NotEmpty(t, user.ID)
}

func TestUserRepo_FindByEmail(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	found, err := repo.FindByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Password, found.Password)
}

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)

	_, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUserRepo_FindByID(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)

	user := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "hashedpassword",
	}

	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)

	_, err := repo.FindByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUserRepo_Create_DuplicateEmail(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewUserRepository(tdb.DB)

	user1 := &domain.User{
		ID:       uuid.New(),
		Email:    "duplicate@example.com",
		Password: "password1",
	}

	user2 := &domain.User{
		ID:       uuid.New(),
		Email:    "duplicate@example.com",
		Password: "password2",
	}

	err := repo.Create(context.Background(), user1)
	require.NoError(t, err)

	err = repo.Create(context.Background(), user2)
	assert.ErrorIs(t, err, domain.ErrConflict)
}
