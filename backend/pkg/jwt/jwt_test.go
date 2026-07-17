package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAccessToken(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		AccessSecret: "test-access-secret-key",
		AccessExpiry: 15 * time.Minute,
	}

	token, err := GenerateAccessToken(userID, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		RefreshSecret: "test-refresh-secret-key",
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	token, err := GenerateRefreshToken(userID, cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateAccessToken(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		AccessSecret: "test-access-secret-key",
		AccessExpiry: 15 * time.Minute,
	}

	token, err := GenerateAccessToken(userID, cfg)
	require.NoError(t, err)

	claims, err := ValidateToken(token, cfg.AccessSecret, TokenTypeAccess)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, TokenTypeAccess, claims.Type)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestValidateRefreshToken(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		RefreshSecret: "test-refresh-secret-key",
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	token, err := GenerateRefreshToken(userID, cfg)
	require.NoError(t, err)

	claims, err := ValidateToken(token, cfg.RefreshSecret, TokenTypeRefresh)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, TokenTypeRefresh, claims.Type)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestValidateToken_Expired(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		AccessSecret: "test-access-secret-key",
		AccessExpiry: -1 * time.Minute, // Already expired
	}

	token, err := GenerateAccessToken(userID, cfg)
	require.NoError(t, err)

	_, err = ValidateToken(token, cfg.AccessSecret, TokenTypeAccess)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestValidateToken_WrongType(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		AccessSecret:  "test-access-secret-key",
		RefreshSecret: "test-refresh-secret-key",
		AccessExpiry:  15 * time.Minute,
	}

	// Generate access token
	token, err := GenerateAccessToken(userID, cfg)
	require.NoError(t, err)

	// Try to validate as refresh token
	_, err = ValidateToken(token, cfg.RefreshSecret, TokenTypeRefresh)
	assert.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		AccessSecret: "test-access-secret-key",
		AccessExpiry: 15 * time.Minute,
	}

	token, err := GenerateAccessToken(userID, cfg)
	require.NoError(t, err)

	// Validate with wrong secret
	_, err = ValidateToken(token, "wrong-secret", TokenTypeAccess)
	assert.Error(t, err)
}

func TestValidateToken_Malformed(t *testing.T) {
	_, err := ValidateToken("not-a-valid-token", "secret", TokenTypeAccess)
	assert.Error(t, err)
}

func TestValidateToken_EmptyToken(t *testing.T) {
	_, err := ValidateToken("", "secret", TokenTypeAccess)
	assert.Error(t, err)
}

func TestClaims_ContainsOnlySafeData(t *testing.T) {
	userID := uuid.New()
	cfg := &Config{
		AccessSecret: "test-access-secret-key",
		AccessExpiry: 15 * time.Minute,
	}

	token, err := GenerateAccessToken(userID, cfg)
	require.NoError(t, err)

	claims, err := ValidateToken(token, cfg.AccessSecret, TokenTypeAccess)
	require.NoError(t, err)

	// Claims should only contain subject (user_id), expiration, issued at, and type
	assert.Equal(t, userID.String(), claims.Subject)
	assert.NotEmpty(t, claims.ExpiresAt)
	assert.NotEmpty(t, claims.IssuedAt)
	assert.Equal(t, TokenTypeAccess, claims.Type)
}

func TestGenerateAccessToken_DifferentUsers(t *testing.T) {
	cfg := &Config{
		AccessSecret: "test-access-secret-key",
		AccessExpiry: 15 * time.Minute,
	}

	userID1 := uuid.New()
	userID2 := uuid.New()

	token1, err := GenerateAccessToken(userID1, cfg)
	require.NoError(t, err)

	token2, err := GenerateAccessToken(userID2, cfg)
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2)

	claims1, err := ValidateToken(token1, cfg.AccessSecret, TokenTypeAccess)
	require.NoError(t, err)
	assert.Equal(t, userID1.String(), claims1.Subject)

	claims2, err := ValidateToken(token2, cfg.AccessSecret, TokenTypeAccess)
	require.NoError(t, err)
	assert.Equal(t, userID2.String(), claims2.Subject)
}
