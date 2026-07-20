package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sergiojoz/gym-tracker/configs"
	"github.com/sergiojoz/gym-tracker/internal/handler"
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
	progressRepo := postgres.NewProgressRepository(db)

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
	progressUC := usecase.NewProgressUseCase(progressRepo, prRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUC)
	exerciseHandler := handler.NewExerciseHandler(exerciseUC)
	templateHandler := handler.NewTemplateHandler(templateUC)
	sessionHandler := handler.NewSessionHandler(sessionUC)
	progressHandler := handler.NewProgressHandler(progressUC)
	mediaHandler := handler.NewMediaHandler(cfg.Media.RootDir, cfg.Media.GIFsDir, cfg.Media.ThumbnailsDir)

	// Setup router
	r := handler.SetupRouter(authHandler, exerciseHandler, templateHandler, sessionHandler, progressHandler, mediaHandler, jwtCfg)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
