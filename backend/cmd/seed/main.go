package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"github.com/sergiojoz/gym-tracker/internal/domain"
	"github.com/sergiojoz/gym-tracker/internal/repository/postgres"
)

// exerciseJSON represents the JSON structure of an exercise in the dataset.
type exerciseJSON struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Category         string              `json:"category"`
	BodyPart         string              `json:"body_part"`
	Equipment        string              `json:"equipment"`
	Instructions     map[string]string   `json:"instructions"`
	InstructionSteps map[string][]string `json:"instruction_steps"`
	MuscleGroup      string              `json:"muscle_group"`
	SecondaryMuscles []string            `json:"secondary_muscles"`
	Target           string              `json:"target"`
	Image            string              `json:"image"`
	GIFURL           string              `json:"gif_url"`
	MediaID          string              `json:"media_id"`
	CreatedAt        string              `json:"created_at"`
	Attribution      string              `json:"attribution"`
}

func main() {
	// Load .env file for local development (silently ignored in production/Docker)
	_ = godotenv.Load()

	datasetPath := os.Getenv("DATASET_PATH")
	if datasetPath == "" {
		log.Fatal("DATASET_PATH environment variable is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	mediaRootDir := os.Getenv("MEDIA_ROOT_DIR")
	if mediaRootDir == "" {
		mediaRootDir = "./media"
	}

	// Open dataset file
	data, err := os.ReadFile(datasetPath)
	if err != nil {
		log.Fatalf("failed to read dataset: %v", err)
	}

	// Parse JSON array
	var rawExercises []json.RawMessage
	if err := json.Unmarshal(data, &rawExercises); err != nil {
		log.Fatalf("failed to parse dataset JSON: %v", err)
	}

	log.Printf("found %d exercises in dataset", len(rawExercises))

	// Load optional translations (Spanish exercise names)
	datasetDir := filepath.Dir(datasetPath)
	translationsPath := filepath.Join(datasetDir, "translations.json")
	translations := loadTranslations(translationsPath)
	if translations != nil {
		log.Printf("loaded %d Spanish name translations", len(translations))
	} else {
		log.Println("no translations file found, exercises will have English names only")
	}

	// Map JSON to domain exercises
	exercises := make([]*domain.Exercise, 0, len(rawExercises))
	for i, raw := range rawExercises {
		ex, err := mapExerciseJSON(raw, translations)
		if err != nil {
			log.Printf("warning: skipping exercise %d: %v", i, err)
			continue
		}
		exercises = append(exercises, ex)
	}

	log.Printf("mapped %d exercises", len(exercises))

	// Connect to database
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("connected to database")

	// Bulk upsert exercises
	repo := postgres.NewExerciseRepository(db)
	ctx := context.Background()

	if err := repo.BulkUpsert(ctx, exercises); err != nil {
		log.Fatalf("failed to upsert exercises: %v", err)
	}

	log.Printf("successfully upserted %d exercises", len(exercises))

	// Copy media files if dataset directory contains them
	copyMediaFiles(datasetDir, mediaRootDir)

	log.Println("seed completed successfully")
}

// mapExerciseJSON converts a raw JSON exercise to a domain Exercise,
// using the optional translations map (dataset ID -> Spanish name) for i18n.
func mapExerciseJSON(raw json.RawMessage, translations map[string]string) (*domain.Exercise, error) {
	var j exerciseJSON
	if err := json.Unmarshal(raw, &j); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if j.Name == "" {
		return nil, fmt.Errorf("missing required field: name")
	}
	if j.MuscleGroup == "" {
		return nil, fmt.Errorf("missing required field: muscle_group")
	}

	// Parse ID if provided, otherwise generate one
	var id uuid.UUID
	if j.ID != "" {
		// Dataset uses numeric string IDs, generate UUID from them
		id = uuid.NewSHA1(uuid.NameSpaceURL, []byte(j.ID))
	} else {
		id = uuid.New()
	}

	// Build multi-language names
	names := map[string]string{"en": j.Name}
	if translations != nil {
		if es, ok := translations[j.ID]; ok && es != "" {
			names["es"] = es
		}
	}

	// Build multi-language descriptions from dataset instructions
	descriptions := make(map[string]string)
	if j.Instructions != nil {
		for lang, text := range j.Instructions {
			descriptions[lang] = text
		}
	}

	// Set default difficulty
	difficulty := "beginner"

	return &domain.Exercise{
		ID:                 id,
		NameByLang:         names,
		DescriptionsByLang: descriptions,
		MuscleGroup:        j.MuscleGroup,
		Equipment:          j.Equipment,
		Difficulty:         difficulty,
		Category:           j.Category,
		GIFPath:            j.GIFURL,
		ThumbnailPath:      j.Image,
	}, nil
}

// loadTranslations reads a JSON file mapping dataset IDs to Spanish names.
// Returns nil if the file does not exist.
func loadTranslations(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		log.Printf("warning: failed to read translations: %v", err)
		return nil
	}
	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		log.Printf("warning: failed to parse translations: %v", err)
		return nil
	}
	return translations
}

// copyMediaFiles copies GIFs and thumbnails from the dataset directory to the media root.
func copyMediaFiles(datasetDir, mediaRootDir string) {
	// Copy GIFs
	gifsSrc := filepath.Join(datasetDir, "gifs")
	gifsDst := filepath.Join(mediaRootDir, "gifs")
	if info, err := os.Stat(gifsSrc); err == nil && info.IsDir() {
		log.Printf("copying GIFs from %s to %s", gifsSrc, gifsDst)
		if err := copyDir(gifsSrc, gifsDst); err != nil {
			log.Printf("warning: failed to copy GIFs: %v", err)
		}
	}

	// Copy thumbnails
	thumbsSrc := filepath.Join(datasetDir, "thumbnails")
	thumbsDst := filepath.Join(mediaRootDir, "thumbnails")
	if info, err := os.Stat(thumbsSrc); err == nil && info.IsDir() {
		log.Printf("copying thumbnails from %s to %s", thumbsSrc, thumbsDst)
		if err := copyDir(thumbsSrc, thumbsDst); err != nil {
			log.Printf("warning: failed to copy thumbnails: %v", err)
		}
	}
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}
