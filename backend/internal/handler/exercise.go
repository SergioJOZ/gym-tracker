package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
)

// ExerciseUseCaseInterface defines the interface for exercise operations.
type ExerciseUseCaseInterface interface {
	List(ctx context.Context, filter repository.ExerciseFilter) ([]*domain.Exercise, bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Exercise, error)
}

// ExerciseHandler handles exercise catalog HTTP requests.
type ExerciseHandler struct {
	exerciseUC ExerciseUseCaseInterface
}

// NewExerciseHandler creates a new ExerciseHandler.
func NewExerciseHandler(exerciseUC ExerciseUseCaseInterface) *ExerciseHandler {
	return &ExerciseHandler{exerciseUC: exerciseUC}
}

// exerciseListResponse represents the paginated list response.
type exerciseListResponse struct {
	Items      []*domain.Exercise `json:"items"`
	NextCursor string             `json:"next_cursor,omitempty"`
	HasMore    bool               `json:"has_more"`
}

// List handles GET /api/v1/exercises.
// Query params: search, muscle_group, equipment, difficulty, category, cursor, limit
func (h *ExerciseHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := repository.ExerciseFilter{
		Search:      q.Get("search"),
		MuscleGroup: q.Get("muscle_group"),
		Equipment:   q.Get("equipment"),
		Difficulty:  q.Get("difficulty"),
		Category:    q.Get("category"),
		Cursor:      q.Get("cursor"),
	}

	// Parse limit
	limitStr := q.Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}

	exercises, hasMore, err := h.exerciseUC.List(r.Context(), filter)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := exerciseListResponse{
		Items:   exercises,
		HasMore: hasMore,
	}

	// Generate next cursor if there are more results
	if hasMore && len(exercises) > 0 {
		last := exercises[len(exercises)-1]
		resp.NextCursor = cursor.Encode(last.ID.String())
	}

	// Ensure items is never null in JSON
	if resp.Items == nil {
		resp.Items = []*domain.Exercise{}
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetByID handles GET /api/v1/exercises/{id}.
func (h *ExerciseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid exercise ID format", http.StatusBadRequest))
		return
	}

	exercise, err := h.exerciseUC.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, exercise)
}
