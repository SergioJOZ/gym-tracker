package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/middlewares"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
)

// TemplateUseCaseInterface defines the interface for template operations.
type TemplateUseCaseInterface interface {
	Create(ctx context.Context, t *domain.WorkoutTemplate) error
	Update(ctx context.Context, t *domain.WorkoutTemplate) error
	Delete(ctx context.Context, userID, templateID uuid.UUID) error
	GetByID(ctx context.Context, userID, templateID uuid.UUID) (*domain.WorkoutTemplate, error)
	List(ctx context.Context, userID uuid.UUID, filter repository.TemplateFilter) ([]*domain.WorkoutTemplate, bool, error)
}

// TemplateHandler handles workout template HTTP requests.
type TemplateHandler struct {
	templateUC TemplateUseCaseInterface
}

// NewTemplateHandler creates a new TemplateHandler.
func NewTemplateHandler(templateUC TemplateUseCaseInterface) *TemplateHandler {
	return &TemplateHandler{templateUC: templateUC}
}

// templateRequest represents the JSON request body for creating/updating a template.
type templateRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Slots       []slotRequest `json:"slots"`
}

// slotRequest represents a single slot in the request.
type slotRequest struct {
	ExerciseID     string   `json:"exercise_id"`
	Order          int      `json:"order"`
	TargetSets     int      `json:"target_sets"`
	TargetReps     int      `json:"target_reps"`
	TargetWeight   *float64 `json:"target_weight,omitempty"`
	TargetDuration *int     `json:"target_duration,omitempty"`
}

// templateListResponse represents the paginated list response.
type templateListResponse struct {
	Items      []*domain.WorkoutTemplate `json:"items"`
	NextCursor string                    `json:"next_cursor,omitempty"`
	HasMore    bool                      `json:"has_more"`
}

// Create handles POST /api/v1/templates.
func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	var req templateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	tmpl := h.requestToDomain(userID, &req)

	if err := h.templateUC.Create(r.Context(), tmpl); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, tmpl)
}

// List handles GET /api/v1/templates.
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	cursorStr, limit := ParsePaginationParams(r)

	filter := repository.TemplateFilter{
		Cursor: cursorStr,
		Limit:  limit,
	}

	templates, hasMore, err := h.templateUC.List(r.Context(), userID, filter)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := templateListResponse{
		Items:   templates,
		HasMore: hasMore,
	}

	if hasMore && len(templates) > 0 {
		last := templates[len(templates)-1]
		resp.NextCursor = cursor.Encode(last.CreatedAt.String(), last.ID.String())
	}

	if resp.Items == nil {
		resp.Items = []*domain.WorkoutTemplate{}
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetByID handles GET /api/v1/templates/{id}.
func (h *TemplateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid template ID format", http.StatusBadRequest))
		return
	}

	tmpl, err := h.templateUC.GetByID(r.Context(), userID, id)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, tmpl)
}

// Update handles PUT /api/v1/templates/{id}.
func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid template ID format", http.StatusBadRequest))
		return
	}

	var req templateRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	tmpl := h.requestToDomain(userID, &req)
	tmpl.ID = id

	if err := h.templateUC.Update(r.Context(), tmpl); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, tmpl)
}

// Delete handles DELETE /api/v1/templates/{id}.
func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid template ID format", http.StatusBadRequest))
		return
	}

	if err := h.templateUC.Delete(r.Context(), userID, id); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// requestToDomain converts a templateRequest to a domain.WorkoutTemplate.
func (h *TemplateHandler) requestToDomain(userID uuid.UUID, req *templateRequest) *domain.WorkoutTemplate {
	tmpl := &domain.WorkoutTemplate{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	}

	for _, s := range req.Slots {
		slot := &domain.TemplateSlot{
			ExerciseID:     uuid.MustParse(s.ExerciseID),
			Order:          s.Order,
			TargetSets:     s.TargetSets,
			TargetReps:     s.TargetReps,
			TargetWeight:   s.TargetWeight,
			TargetDuration: s.TargetDuration,
		}
		tmpl.Slots = append(tmpl.Slots, slot)
	}

	return tmpl
}
