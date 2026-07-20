package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// MediaHandler serves static media files (GIFs and thumbnails).
type MediaHandler struct {
	rootDir       string
	gifsDir       string
	thumbnailsDir string
}

// NewMediaHandler creates a new MediaHandler.
func NewMediaHandler(rootDir, gifsDir, thumbnailsDir string) *MediaHandler {
	return &MediaHandler{
		rootDir:       rootDir,
		gifsDir:       gifsDir,
		thumbnailsDir: thumbnailsDir,
	}
}

// ServeGIF handles GET /media/gifs/{filename}.
func (h *MediaHandler) ServeGIF(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	h.serveFile(w, r, h.gifsDir, filename)
}

// ServeThumbnail handles GET /media/thumbnails/{filename}.
func (h *MediaHandler) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	h.serveFile(w, r, h.thumbnailsDir, filename)
}

// serveFile serves a single file from the given subdirectory, with path traversal protection.
func (h *MediaHandler) serveFile(w http.ResponseWriter, r *http.Request, subDir, filename string) {
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Prevent path traversal: reject any filename containing path separators or ".."
	if strings.ContainsAny(filename, "/\\") || strings.Contains(filename, "..") {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Build the full path and verify it stays within the expected directory
	fullPath := filepath.Join(h.rootDir, subDir, filename)
	expectedDir := filepath.Join(h.rootDir, subDir)

	// Clean and verify the resolved path is within the expected directory
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(expectedDir)+string(filepath.Separator)) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Verify the file exists and is a regular file
	stat, err := os.Stat(cleanPath)
	if err != nil || stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, cleanPath)
}
