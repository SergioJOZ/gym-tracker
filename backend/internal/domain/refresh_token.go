package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token stored in the database.
type RefreshToken struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	TokenHash   string     `json:"-"` // hashed token for lookup
	ExpiresAt   time.Time  `json:"expires_at"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"` // NULL = active
	CreatedAt   time.Time  `json:"created_at"`
}

// IsRevoked returns true if the token has been revoked.
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsExpired returns true if the token has expired.
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid returns true if the token is both not revoked and not expired.
func (t *RefreshToken) IsValid() bool {
	return !t.IsRevoked() && !t.IsExpired()
}
