package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/middlewares"
	"github.com/sergiojoz/gym-tracker/internal/repository"
	"github.com/sergiojoz/gym-tracker/pkg/cursor"
)

// SessionUseCaseInterface defines the interface for session operations.
type SessionUseCaseInterface interface {
	Create(ctx context.Context, s *domain.WorkoutSession) error
	Update(ctx context.Context, s *domain.WorkoutSession) error
	Delete(ctx context.Context, userID, sessionID uuid.UUID) error
	GetByID(ctx context.Context, userID, sessionID uuid.UUID) (*domain.WorkoutSession, error)
	List(ctx context.Context, userID uuid.UUID, filter repository.SessionFilter) ([]*domain.WorkoutSession, bool, error)
}

// SessionHandler handles workout session HTTP requests.
type SessionHandler struct {
	sessionUC SessionUseCaseInterface
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessionUC SessionUseCaseInterface) *SessionHandler {
	return &SessionHandler{sessionUC: sessionUC}
}

// sessionRequest represents the JSON request body for creating/updating a session.
type sessionRequest struct {
	Name        string               `json:"name"`
	Notes       string               `json:"notes"`
	TemplateID  *string              `json:"template_id,omitempty"`
	StartAt     string               `json:"start_at"`
	EndAt       *string              `json:"end_at,omitempty"`
	Exercises   []sessionExerciseReq `json:"exercises"`
}

// sessionExerciseReq represents a single exercise in the request.
type sessionExerciseReq struct {
	ExerciseID string          `json:"exercise_id"`
	Order      int             `json:"order"`
	Notes      string          `json:"notes"`
	Sets       []sessionSetReq `json:"sets"`
}

// sessionSetReq represents a single set in the request.
type sessionSetReq struct {
	Order    int      `json:"order"`
	Weight   *float64 `json:"weight,omitempty"`
	Reps     *int     `json:"reps,omitempty"`
	Duration *int     `json:"duration,omitempty"`
	RPE      *float64 `json:"rpe,omitempty"`
}

// sessionListResponse represents the paginated list response.
type sessionListResponse struct {
	Items      []*domain.WorkoutSession `json:"items"`
	NextCursor string                   `json:"next_cursor,omitempty"`
	HasMore    bool                     `json:"has_more"`
}

// Create handles POST /api/v1/sessions.
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	var req sessionRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	session, err := h.requestToDomain(userID, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	if err := h.sessionUC.Create(r.Context(), session); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, session)
}

// List handles GET /api/v1/sessions.
func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	cursorStr, limit := ParsePaginationParams(r)

	filter := repository.SessionFilter{
		Cursor: cursorStr,
		Limit:  limit,
	}

	// Parse optional date filters
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &t
		}
	}

	sessions, hasMore, err := h.sessionUC.List(r.Context(), userID, filter)
	if err != nil {
		respondError(w, err)
		return
	}

	resp := sessionListResponse{
		Items:   sessions,
		HasMore: hasMore,
	}

	if hasMore && len(sessions) > 0 {
		last := sessions[len(sessions)-1]
		resp.NextCursor = cursor.Encode(last.StartAt.Format(time.RFC3339Nano), last.ID.String())
	}

	if resp.Items == nil {
		resp.Items = []*domain.WorkoutSession{}
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetByID handles GET /api/v1/sessions/{id}.
func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid session ID format", http.StatusBadRequest))
		return
	}

	session, err := h.sessionUC.GetByID(r.Context(), userID, id)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// Update handles PUT /api/v1/sessions/{id}.
func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid session ID format", http.StatusBadRequest))
		return
	}

	var req sessionRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		respondError(w, err)
		return
	}

	session, err := h.requestToDomain(userID, &req)
	if err != nil {
		respondError(w, err)
		return
	}
	session.ID = id

	if err := h.sessionUC.Update(r.Context(), session); err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// Delete handles DELETE /api/v1/sessions/{id}.
func (h *SessionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, domain.ErrUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, domain.NewAppError("INVALID_ID", "invalid session ID format", http.StatusBadRequest))
		return
	}

	if err := h.sessionUC.Delete(r.Context(), userID, id); err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// requestToDomain converts a sessionRequest to a domain.WorkoutSession.
func (h *SessionHandler) requestToDomain(userID uuid.UUID, req *sessionRequest) (*domain.WorkoutSession, error) {
	session := &domain.WorkoutSession{
		UserID: userID,
		Name:   req.Name,
		Notes:  req.Notes,
	}

	// Parse start_at
	if req.StartAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, domain.NewAppError("INVALID_DATE", "invalid start_at format", http.StatusBadRequest)
		}
		session.StartAt = t
	} else {
		session.StartAt = time.Now()
	}

	// Parse end_at
	if req.EndAt != nil && *req.EndAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EndAt)
		if err != nil {
			return nil, domain.NewAppError("INVALID_DATE", "invalid end_at format", http.StatusBadRequest)
		}
		session.EndAt = &t
	}

	// Parse template_id
	if req.TemplateID != nil && *req.TemplateID != "" {
		tid, err := uuid.Parse(*req.TemplateID)
		if err != nil {
			return nil, domain.NewAppError("INVALID_ID", "invalid template_id format", http.StatusBadRequest)
		}
		session.TemplateID = &tid
	}

	// Parse exercises and sets
	for _, exReq := range req.Exercises {
		ex := &domain.SessionExercise{
			ExerciseID: uuid.MustParse(exReq.ExerciseID),
			Order:      exReq.Order,
			Notes:      exReq.Notes,
		}

		for _, setReq := range exReq.Sets {
			set := &domain.SessionSet{
				Order:    setReq.Order,
				Weight:   setReq.Weight,
				Reps:     setReq.Reps,
				Duration: setReq.Duration,
				RPE:      setReq.RPE,
			}
			ex.Sets = append(ex.Sets, set)
		}

		session.Exercises = append(session.Exercises, ex)
	}

	return session, nil
}
