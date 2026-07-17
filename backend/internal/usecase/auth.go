package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	jwtPkg "github.com/sergiojoz/gym-tracker/pkg/jwt"
	"github.com/sergiojoz/gym-tracker/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase handles authentication business logic.
type AuthUseCase struct {
	userRepo      repository.UserRepository
	tokenRepo     repository.RefreshTokenRepository
	jwtCfg        *jwtPkg.Config
}

// NewAuthUseCase creates a new AuthUseCase.
func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	jwtCfg *jwtPkg.Config,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtCfg:    jwtCfg,
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

// RegisterResponse represents the response after successful registration.
type RegisterResponse struct {
	User *domain.User
}

// Register creates a new user account.
func (uc *AuthUseCase) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	// Validate input
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	_, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, domain.ErrConflict
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResponse{User: user}, nil
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

// LoginResponse represents the response after successful login.
type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}

// Login authenticates a user and returns tokens.
func (uc *AuthUseCase) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Validate input
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	// Find user by email
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal whether email exists
		return nil, domain.ErrUnauthorized
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Generate access token
	accessToken, err := jwtPkg.GenerateAccessToken(user.ID, uc.jwtCfg)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := uc.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	RefreshToken string `validate:"required"`
}

// RefreshResponse represents the response after successful token refresh.
type RefreshResponse struct {
	AccessToken  string
	RefreshToken string
}

// Refresh rotates the refresh token and issues a new access token.
func (uc *AuthUseCase) Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResponse, error) {
	// Validate input
	if err := validator.Validate(req); err != nil {
		return nil, err
	}

	// Hash the refresh token to look it up
	tokenHash := hashToken(req.RefreshToken)

	// Find the token in the database
	storedToken, err := uc.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Check if token is valid
	if !storedToken.IsValid() {
		return nil, domain.ErrUnauthorized
	}

	// Revoke old token
	if err := uc.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := jwtPkg.GenerateAccessToken(storedToken.UserID, uc.jwtCfg)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := uc.generateRefreshToken(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	RefreshToken string `validate:"required"`
}

// Logout revokes the refresh token.
func (uc *AuthUseCase) Logout(ctx context.Context, req *LogoutRequest) error {
	// Validate input
	if err := validator.Validate(req); err != nil {
		return err
	}

	// Hash the refresh token to look it up
	tokenHash := hashToken(req.RefreshToken)

	// Find the token in the database
	storedToken, err := uc.tokenRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrUnauthorized
	}

	// Revoke the token
	return uc.tokenRepo.Revoke(ctx, storedToken.ID)
}

// generateRefreshToken creates a new refresh token, stores its hash, and returns the raw token.
func (uc *AuthUseCase) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Generate random token
	rawToken := generateRandomToken()
	tokenHash := hashToken(rawToken)

	// Store token hash in database
	token := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().Add(uc.jwtCfg.RefreshExpiry),
	}

	if err := uc.tokenRepo.Create(ctx, token); err != nil {
		return "", err
	}

	return rawToken, nil
}

// generateRandomToken generates a cryptographically secure random token.
func generateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // This should never happen
	}
	return hex.EncodeToString(b)
}

// hashToken creates a SHA-256 hash of the token for storage.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
