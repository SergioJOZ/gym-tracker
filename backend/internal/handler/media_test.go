package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaHandler_ServeGIF(t *testing.T) {
	// Create temp directory with a test GIF
	tmpDir := t.TempDir()
	gifsDir := filepath.Join(tmpDir, "gifs")
	require.NoError(t, os.MkdirAll(gifsDir, 0755))

	// Create a test file
	testContent := []byte("fake gif content")
	require.NoError(t, os.WriteFile(filepath.Join(gifsDir, "bench_press.gif"), testContent, 0644))

	handler := NewMediaHandler(tmpDir, "gifs", "thumbnails")

	req := httptest.NewRequest(http.MethodGet, "/media/gifs/bench_press.gif", nil)
	req = setURLParam(req, "filename", "bench_press.gif")
	w := httptest.NewRecorder()

	handler.ServeGIF(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testContent, w.Body.Bytes())
}

func TestMediaHandler_ServeThumbnail(t *testing.T) {
	// Create temp directory with a test thumbnail
	tmpDir := t.TempDir()
	thumbnailsDir := filepath.Join(tmpDir, "thumbnails")
	require.NoError(t, os.MkdirAll(thumbnailsDir, 0755))

	// Create a test file
	testContent := []byte("fake thumbnail content")
	require.NoError(t, os.WriteFile(filepath.Join(thumbnailsDir, "bench_press.jpg"), testContent, 0644))

	handler := NewMediaHandler(tmpDir, "gifs", "thumbnails")

	req := httptest.NewRequest(http.MethodGet, "/media/thumbnails/bench_press.jpg", nil)
	req = setURLParam(req, "filename", "bench_press.jpg")
	w := httptest.NewRecorder()

	handler.ServeThumbnail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testContent, w.Body.Bytes())
}

func TestMediaHandler_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	gifsDir := filepath.Join(tmpDir, "gifs")
	require.NoError(t, os.MkdirAll(gifsDir, 0755))

	handler := NewMediaHandler(tmpDir, "gifs", "thumbnails")

	req := httptest.NewRequest(http.MethodGet, "/media/gifs/nonexistent.gif", nil)
	req = setURLParam(req, "filename", "nonexistent.gif")
	w := httptest.NewRecorder()

	handler.ServeGIF(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMediaHandler_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	gifsDir := filepath.Join(tmpDir, "gifs")
	require.NoError(t, os.MkdirAll(gifsDir, 0755))

	handler := NewMediaHandler(tmpDir, "gifs", "thumbnails")

	// Try to access file outside the gifs directory
	req := httptest.NewRequest(http.MethodGet, "/media/gifs/../../../etc/passwd", nil)
	req = setURLParam(req, "filename", "../../../etc/passwd")
	w := httptest.NewRecorder()

	handler.ServeGIF(w, req)

	// Should return 404 or 400, not serve the file
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest)
}
