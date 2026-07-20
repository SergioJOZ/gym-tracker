package cursor

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultLimit is the default page size when not specified.
	DefaultLimit = 20
	// MaxLimit is the maximum allowed page size.
	MaxLimit = 100
	// separator is used to join cursor values before encoding.
	separator = "\x1F" // ASCII Unit Separator
)

// PageRequest represents a cursor-based pagination request.
type PageRequest struct {
	Cursor string // opaque cursor, empty = first page
	Limit  int    // page size, 0 or negative = DefaultLimit
}

// Page represents a paginated response.
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// Validate normalizes the PageRequest, applying defaults and limits.
func (p PageRequest) Validate() PageRequest {
	if p.Limit <= 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	return p
}

// Encode creates an opaque cursor from the given values.
// Values are joined with a separator and base64-encoded.
func Encode(values ...string) string {
	joined := strings.Join(values, separator)
	return base64.StdEncoding.EncodeToString([]byte(joined))
}

// Decode decodes an opaque cursor back into its component values.
// expectedCount specifies how many values are expected.
func Decode(cursor string, expectedCount int) ([]string, error) {
	// Handle empty cursor
	if cursor == "" {
		if expectedCount == 0 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("invalid cursor: expected %d fields, got 0", expectedCount)
	}

	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: failed to decode base64: %w", err)
	}

	// Handle empty decoded content
	if len(decoded) == 0 {
		if expectedCount == 0 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("invalid cursor: expected %d fields, got 0", expectedCount)
	}

	parts := strings.Split(string(decoded), separator)

	if len(parts) != expectedCount {
		return nil, fmt.Errorf("invalid cursor: expected %d fields, got %d", expectedCount, len(parts))
	}

	return parts, nil
}

// NewPage creates a Page from a slice of items.
// It fetches limit+1 items to determine HasMore.
// cursorFunc converts the last item to a cursor string.
func NewPage[T any](items []T, limit int, cursorFunc func(T) string) Page[T] {
	hasMore := len(items) > limit
	
	var result []T
	if hasMore {
		result = items[:limit]
	} else {
		result = items
	}

	page := Page[T]{
		Items:   result,
		HasMore: hasMore,
	}

	if hasMore && len(result) > 0 {
		page.NextCursor = cursorFunc(result[len(result)-1])
	}

	return page
}

// ParseInt parses a string to int with a default value.
func ParseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
