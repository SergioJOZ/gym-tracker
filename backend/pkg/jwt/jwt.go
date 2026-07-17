package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType represents the type of JWT token.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Config holds JWT configuration.
type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// Claims represents the JWT claims.
type Claims struct {
	Subject string    `json:"sub"`
	Type    TokenType `json:"type"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a new access token for the given user.
func GenerateAccessToken(userID uuid.UUID, cfg *Config) (string, error) {
	return generateToken(userID, cfg.AccessSecret, cfg.AccessExpiry, TokenTypeAccess)
}

// GenerateRefreshToken creates a new refresh token for the given user.
func GenerateRefreshToken(userID uuid.UUID, cfg *Config) (string, error) {
	return generateToken(userID, cfg.RefreshSecret, cfg.RefreshExpiry, TokenTypeRefresh)
}

// ValidateToken validates a JWT token and returns the claims.
func ValidateToken(tokenString, secret string, expectedType TokenType) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validate token type
	if claims.Type != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.Type)
	}

	return claims, nil
}

// generateToken creates a JWT token with the given parameters.
func generateToken(userID uuid.UUID, secret string, expiry time.Duration, tokenType TokenType) (string, error) {
	now := time.Now()
	claims := &Claims{
		Subject: userID.String(),
		Type:    tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
