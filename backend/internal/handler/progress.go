package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/middlewares"
	"github.com/sergiojoz/gym-tracker/internal/repository"
)

// ProgressUseCaseInterface defines the interface for progress operations.
type ProgressUseCaseInterface interface {
	ListPRs(ctx context.Context, userID uuid.UUID) ([]*domain.PersonalRecord, error)
	ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, cursor string, limit int) ([]*domain.SessionSet, bool, error)
	Summary(ctx context.Context, userID uuid.UUID) (*repository.ProgressSummary, error)
}

// ProgressHandler handles progress and history HTTP requests.
type ProgressHandler struct {
	progressUC ProgressUseCaseInterface
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler(progressUC ProgressUseCaseInterface) *ProgressHandler {
	return &ProgressHandler{progressUC: progressUC}
}

// progressListResponse represents the paginated list response for exercise history.
type progressListResponse struct {
	Items      []*domain.SessionSet `json:"items"`
	NextCursor string               `json:"next_cursor,omitempty"`
	HasMore    bool                 `json:"has_more"`
}

// ListPRs handles GET /api/v1/progress/records.
func (h *ProgressHandler) ListPRs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	prs, err := h.progressUC.ListPRs(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	if prs == nil {
		prs = []*domain.PersonalRecord{}
	}

	respondJSON(w, http.StatusOK, prs)
}

// ExerciseHistory handles GET /api/v1/progress/exercises/{id}/history.
func (h *ProgressHandler) ExerciseHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	exerciseID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid exercise ID format", http.StatusBadRequest))
		return
	}

	cursorStr, limit := ParsePaginationParams(r)

	sets, hasMore, err := h.progressUC.ExerciseHistory(r.Context(), userID, exerciseID, cursorStr, limit)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := progressListResponse{
		Items:   sets,
		HasMore: hasMore,
	}

	if resp.Items == nil {
		resp.Items = []*domain.SessionSet{}
	}

	respondJSON(w, http.StatusOK, resp)
}

// Summary handles GET /api/v1/progress/summary.
func (h *ProgressHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	summary, err := h.progressUC.Summary(r.Context(), userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}
