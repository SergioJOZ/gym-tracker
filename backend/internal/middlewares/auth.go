package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	jwtPkg "github.com/sergiojoz/gym-tracker/pkg/jwt"
)

type contextKey string

const userIDKey contextKey = "user_id"

// AuthMiddleware creates a middleware that validates JWT access tokens.
func AuthMiddleware(jwtCfg *jwtPkg.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			tokenString := extractToken(authHeader)

			if tokenString == "" {
				respondError(w, domain.ErrUnauthorized)
				return
			}

			// Validate token
			claims, err := jwtPkg.ValidateToken(tokenString, jwtCfg.AccessSecret, jwtPkg.TokenTypeAccess)
			if err != nil {
				respondError(w, domain.ErrUnauthorized)
				return
			}

			// Parse user ID from claims
			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				respondError(w, domain.ErrUnauthorized)
				return
			}

			// Inject user ID into context
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

// extractToken extracts the token from the Authorization header.
func extractToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}

	scheme := strings.ToLower(parts[0])
	if scheme != "bearer" {
		return ""
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return ""
	}

	return token
}

// respondError writes an error response.
func respondError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	if e, ok := err.(*domain.AppError); ok {
		appErr = e
	} else {
		appErr = &domain.AppError{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
			Status:  http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	w.Write([]byte(`{"error":{"code":"` + appErr.Code + `","message":"` + appErr.Message + `"}}`))
}
