package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sergiojoz/gym-tracker/configs"
	"github.com/sergiojoz/gym-tracker/internal/handler"
	"github.com/sergiojoz/gym-tracker/internal/middlewares"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres"
	"github.com/sergiojoz/gym-tracker/internal/usecase"
	jwtPkg "github.com/sergiojoz/gym-tracker/pkg/jwt"
)

func main() {
	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("pgx", cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("connected to database")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewRefreshTokenRepository(db)
	exerciseRepo := postgres.NewExerciseRepository(db)
	templateRepo := postgres.NewTemplateRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	prRepo := postgres.NewPersonalRecordRepository(db)

	// Initialize JWT config
	jwtCfg := &jwtPkg.Config{
		AccessSecret:  cfg.JWT.AccessSecret,
		RefreshSecret: cfg.JWT.RefreshSecret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	}

	// Initialize use cases
	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, jwtCfg)
	exerciseUC := usecase.NewExerciseUseCase(exerciseRepo)
	templateUC := usecase.NewTemplateUseCase(templateRepo, exerciseRepo)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, prRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUC)
	exerciseHandler := handler.NewExerciseHandler(exerciseUC)
	templateHandler := handler.NewTemplateHandler(templateUC)
	sessionHandler := handler.NewSessionHandler(sessionUC)
	mediaHandler := handler.NewMediaHandler(cfg.Media.RootDir, cfg.Media.GIFsDir, cfg.Media.ThumbnailsDir)

	// Setup router
	r := setupRouter(authHandler, exerciseHandler, templateHandler, sessionHandler, mediaHandler, jwtCfg)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func setupRouter(
	authHandler *handler.AuthHandler,
	exerciseHandler *handler.ExerciseHandler,
	templateHandler *handler.TemplateHandler,
	sessionHandler *handler.SessionHandler,
	mediaHandler *handler.MediaHandler,
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
		})
	})

	// Media routes (public, no auth required)
	r.Route("/media", func(r chi.Router) {
		r.Get("/gifs/{filename}", mediaHandler.ServeGIF)
		r.Get("/thumbnails/{filename}", mediaHandler.ServeThumbnail)
	})

	return r
}
