# Media Serving Specification

## Purpose

Serve exercise GIFs and thumbnails as static files from the Go backend. Files are stored on disk and served via http.FileServer.

## Requirements

### Requirement: Serve Exercise GIFs

The system MUST serve exercise demonstration GIFs from a configured static directory.

#### Scenario: Request existing GIF

- GIVEN a GIF file exists at media/gifs/{exercise_id}.gif
- WHEN the user requests GET /media/gifs/{exercise_id}.gif
- THEN the system returns 200 with the GIF file and correct Content-Type (image/gif)

#### Scenario: Request non-existing GIF

- GIVEN no GIF file for exercise_id "xyz"
- WHEN the user requests GET /media/gifs/xyz.gif
- THEN the system returns 404

### Requirement: Serve Exercise Thumbnails

The system MUST serve thumbnail images from a configured static directory.

#### Scenario: Request existing thumbnail

- GIVEN a thumbnail file exists at media/thumbnails/{exercise_id}.jpg
- WHEN the user requests GET /media/thumbnails/{exercise_id}.jpg
- THEN the system returns 200 with the image and correct Content-Type

#### Scenario: Request non-existing thumbnail

- GIVEN no thumbnail for exercise_id "xyz"
- WHEN the user requests GET /media/thumbnails/xyz.jpg
- THEN the system returns 404

### Requirement: Media URL References in Exercise Responses

The system MUST include gif_url and thumbnail_url fields in exercise API responses, pointing to the static file paths.

#### Scenario: Exercise response includes media URLs

- GIVEN an exercise with id "abc-123" and associated media files
- WHEN the user requests GET /exercises/abc-123
- THEN the response includes gif_url: "/media/gifs/abc-123.gif" and thumbnail_url: "/media/thumbnails/abc-123.jpg"

## Constraints

- Media files are served directly from Go backend (no CDN in MVP)
- Static file directory path MUST be configurable via environment variable
- GIF files SHOULD be cached by the browser (Cache-Control headers)
- Directory traversal attacks MUST be prevented (no ../ in paths)
- Media endpoints MAY be unauthenticated (public access to exercise media)

## Dependencies

- exercise-catalog (exercise IDs reference media files)
