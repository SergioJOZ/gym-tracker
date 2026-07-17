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

	// Initialize JWT config
	jwtCfg := &jwtPkg.Config{
		AccessSecret:  cfg.JWT.AccessSecret,
		RefreshSecret: cfg.JWT.RefreshSecret,
		AccessExpiry:  cfg.JWT.AccessExpiry,
		RefreshExpiry: cfg.JWT.RefreshExpiry,
	}

	// Initialize use cases
	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, jwtCfg)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUC)

	// Setup router
	r := setupRouter(authHandler, jwtCfg)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func setupRouter(authHandler *handler.AuthHandler, jwtCfg *jwtPkg.Config) *chi.Mux {
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

			// Add protected routes here as they are implemented
			// For now, just a health check
			r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			})
		})
	})

	return r
}
