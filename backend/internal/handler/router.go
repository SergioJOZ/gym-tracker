package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sergiojoz/gym-tracker/internal/middlewares"
	jwtPkg "github.com/sergiojoz/gym-tracker/pkg/jwt"
)

// SetupRouter creates and configures the chi router with all routes and middleware.
func SetupRouter(
	authHandler *AuthHandler,
	exerciseHandler *ExerciseHandler,
	templateHandler *TemplateHandler,
	sessionHandler *SessionHandler,
	progressHandler *ProgressHandler,
	mediaHandler *MediaHandler,
	jwtCfg *jwtPkg.Config,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware chain
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Public routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)
		r.Post("/auth/logout", authHandler.Logout)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(jwtCfg))

			// Health check
			r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			})

			// Exercise routes (protected)
			r.Get("/exercises", exerciseHandler.List)
			r.Get("/exercises/{id}", exerciseHandler.GetByID)

			// Template routes (protected)
			r.Post("/templates", templateHandler.Create)
			r.Get("/templates", templateHandler.List)
			r.Get("/templates/{id}", templateHandler.GetByID)
			r.Put("/templates/{id}", templateHandler.Update)
			r.Delete("/templates/{id}", templateHandler.Delete)

			// Session routes (protected)
			r.Post("/sessions", sessionHandler.Create)
			r.Get("/sessions", sessionHandler.List)
			r.Get("/sessions/{id}", sessionHandler.GetByID)
			r.Put("/sessions/{id}", sessionHandler.Update)
			r.Delete("/sessions/{id}", sessionHandler.Delete)

			// Progress routes (protected)
			r.Get("/progress/records", progressHandler.ListPRs)
			r.Get("/progress/exercises/{id}/history", progressHandler.ExerciseHistory)
			r.Get("/progress/summary", progressHandler.Summary)
		})
	})

	// Media routes (public, no auth required)
	r.Route("/media", func(r chi.Router) {
		r.Get("/gifs/{filename}", mediaHandler.ServeGIF)
		r.Get("/thumbnails/{filename}", mediaHandler.ServeThumbnail)
	})

	return r
}
