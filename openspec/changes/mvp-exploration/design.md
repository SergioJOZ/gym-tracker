# Design: Gym-Tracker MVP

## Technical Approach

Monolithic Go binary with clean/hexagonal architecture. Seven capabilities (auth, exercise-catalog, workout-templates, workout-sessions, personal-records, progress, media-serving) implemented as domain-focused packages. PostgreSQL with cursor-based pagination throughout. Exercise data seeded from `exercises-dataset` JSON (1,324 records) via CLI command. Media served via `http.FileServer`. Strict TDD: `go test -race -cover ./...` with ≥80% coverage.

## Architecture Decisions

| # | Decision | Choice | Alternatives | Rationale |
|---|----------|--------|--------------|-----------|
| D1 | ID strategy | UUID v7 (time-sortable) | UUID v4, auto-increment | Time-sortable = natural cursor for pagination; no coordination needed; globally unique |
| D2 | Pagination | Cursor-based (keyset) on all list endpoints | Offset-based | Consistent results under concurrent inserts; better perf on large datasets |
| D3 | Exercise data model | Flat table with `target_muscles` as `text[]` | JSONB for muscles | Array supports GIN index for filtering; simpler queries; no JSON parsing overhead |
| D4 | PR calculation | Inline within session save transaction | Background job / event | Atomicity guarantee; no eventual consistency issues for MVP |
| D5 | Template deletion | Hard delete (cascade to slots only) | Soft delete | Simpler for MVP; sessions reference exercises directly, not templates |
| D6 | Refresh token storage | DB table with `revoked_at` column | JWT denylist / Redis | Single-tenant MVP; DB is sufficient; supports logout + rotation |
| D7 | Router | `chi` (stdlib-compatible) | gin, echo, fiber | Minimal deps; stdlib `net/http` handlers; easy to test; middleware chaining |
| D8 | Migrations | `golang-migrate/migrate` | goose, atlas | Already referenced in Makefile; simple up/down; SQL files |
| D9 | Config | Env vars + optional `.env` file (caarlos0/env) | YAML config, viper | 12-factor; simple; Docker-friendly |
| D10 | Exercise search | PostgreSQL `tsvector` + GIN index | Application-level search | Native FTS; <100ms on 1,324 rows; no external deps |

## Data Flow

```
HTTP Request
    │
    ▼
┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌────────────┐
│  Middleware   │────▶│   Handler    │────▶│   UseCase    │────▶│ Repository │
│ (auth,log,   │     │ (decode req, │     │ (business    │     │ (SQL,      │
│  error)      │     │  encode resp)│     │  logic, PR   │     │  tx,       │
│              │     │              │     │  calc)       │     │  cursor)   │
└─────────────┘     └──────────────┘     └──────────────┘     └────────────┘
                                                                    │
                                                                    ▼
                                                              ┌────────────┐
                                                              │ PostgreSQL │
                                                              └────────────┘
```

**PR recalculation flow (on session create/update):**
```
Handler ──▶ UseCase.CreateSession()
              │
              ├── BEGIN TX
              ├── INSERT session + exercises + sets
              ├── FOR EACH exercise in session:
              │     compute session_max_weight, session_max_reps, session_volume
              │     UPSERT personal_records (GREATEST of existing vs new)
              └── COMMIT TX
```

**PR recalculation flow (on session delete):**
```
Handler ──▶ UseCase.DeleteSession()
              │
              ├── BEGIN TX
              ├── Fetch deleted session's exercises + sets
              ├── DELETE session (cascade to exercises/sets)
              ├── FOR EACH affected exercise:
              │     IF deleted session held a PR value:
              │       recompute from remaining sessions (SELECT MAX...)
              │       UPDATE personal_records
              └── COMMIT TX
```

## Package Structure

```
backend/
├── cmd/
│   ├── api/
│   │   └── main.go              # App bootstrap, DI wiring, router setup
│   └── seed/
│       └── main.go              # Exercise seeder CLI
├── internal/
│   ├── domain/
│   │   ├── user.go              # User entity
│   │   ├── exercise.go          # Exercise entity (read-only)
│   │   ├── template.go          # WorkoutTemplate + TemplateSlot
│   │   ├── session.go           # WorkoutSession + SessionExercise + SessionSet
│   │   ├── personal_record.go   # PersonalRecord entity
│   │   └── errors.go            # Domain error types (ErrNotFound, ErrConflict, etc.)
│   ├── usecase/
│   │   ├── auth.go              # Register, Login, Refresh, Logout
│   │   ├── auth_test.go
│   │   ├── exercise.go          # List, Search, Filter, GetByID
│   │   ├── exercise_test.go
│   │   ├── template.go          # CRUD templates
│   │   ├── template_test.go
│   │   ├── session.go           # CRUD sessions + PR trigger
│   │   ├── session_test.go
│   │   ├── progress.go          # ListPRs, ExerciseHistory, Summary
│   │   └── progress_test.go
│   ├── repository/
│   │   ├── interfaces.go        # All repository interfaces (ports)
│   │   └── postgres/
│   │       ├── user.go
│   │       ├── user_test.go     # Integration tests (test DB)
│   │       ├── exercise.go
│   │       ├── exercise_test.go
│   │       ├── template.go
│   │       ├── template_test.go
│   │       ├── session.go
│   │       ├── session_test.go
│   │       ├── personal_record.go
│   │       ├── personal_record_test.go
│   │       ├── progress.go
│   │       ├── progress_test.go
│   │       └── testutil/
│   │           └── db.go        # Test DB setup/teardown helpers
│   ├── handler/
│   │   ├── router.go            # Route registration, middleware chain
│   │   ├── auth.go              # POST /auth/register, login, refresh, logout
│   │   ├── auth_test.go
│   │   ├── exercise.go          # GET /exercises, /exercises/{id}
│   │   ├── exercise_test.go
│   │   ├── template.go          # CRUD /templates
│   │   ├── template_test.go
│   │   ├── session.go           # CRUD /sessions
│   │   ├── session_test.go
│   │   ├── progress.go          # GET /progress/records, exercises/{id}/history, summary
│   │   ├── progress_test.go
│   │   ├── media.go             # Static file serving setup
│   │   ├── response.go          # JSON response helpers, error formatting
│   │   └── request.go           # Request decoding, pagination param parsing
│   └── middlewares/
│       ├── auth.go              # JWT extraction + validation
│       ├── auth_test.go
│       ├── logger.go            # Request logging
│       └── recovery.go          # Panic recovery
├── pkg/
│   ├── validator/
│   │   ├── validator.go         # Input validation (go-playground/validator)
│   │   └── validator_test.go
│   ├── cursor/
│   │   ├── cursor.go            # Cursor encode/decode, pagination helpers
│   │   └── cursor_test.go
│   ├── jwt/
│   │   ├── jwt.go               # Token generation + validation
│   │   └── jwt_test.go
│   └── pagination/
│       ├── pagination.go        # CursorPage request/response types
│       └── pagination_test.go
├── migrations/
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_exercises.up.sql
│   ├── 000002_create_exercises.down.sql
│   ├── 000003_create_templates.up.sql
│   ├── 000003_create_templates.down.sql
│   ├── 000004_create_sessions.up.sql
│   ├── 000004_create_sessions.down.sql
│   ├── 000005_create_personal_records.up.sql
│   ├── 000005_create_personal_records.down.sql
│   ├── 000006_create_refresh_tokens.up.sql
│   └── 000006_create_refresh_tokens.down.sql
├── configs/
│   └── config.go                # Config struct + env loading
├── media/                       # Static media directory (gitignored)
│   ├── gifs/                    # Exercise GIFs (symlinked or copied from dataset)
│   └── thumbnails/              # Exercise thumbnails
├── go.mod
├── go.sum
├── Makefile
├── docker-compose.yml
└── Dockerfile
```

## Database Schema

### Table: `users`
```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,  -- bcrypt hash
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_email ON users(email);
```

### Table: `refresh_tokens`
```sql
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,  -- hashed token for lookup
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,                   -- NULL = active
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
```

### Table: `exercises`
```sql
CREATE TABLE exercises (
    id                VARCHAR(10) PRIMARY KEY,  -- dataset ID "0001"-"5201"
    name              VARCHAR(255) NOT NULL,
    category          VARCHAR(50) NOT NULL,
    body_part         VARCHAR(50) NOT NULL,
    equipment         VARCHAR(50) NOT NULL,
    target_muscle     VARCHAR(50) NOT NULL,     -- "target" field from dataset
    muscle_group      VARCHAR(50) NOT NULL,
    secondary_muscles TEXT[] NOT NULL DEFAULT '{}',
    instructions      TEXT,                     -- English instructions (joined from steps)
    gif_url           VARCHAR(500),
    thumbnail_url     VARCHAR(500),
    search_vector     tsvector,                 -- generated column for FTS
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Full-text search index
CREATE INDEX idx_exercises_search ON exercises USING GIN(search_vector);
-- Filter indexes
CREATE INDEX idx_exercises_category ON exercises(category);
CREATE INDEX idx_exercises_body_part ON exercises(body_part);
CREATE INDEX idx_exercises_equipment ON exercises(equipment);
CREATE INDEX idx_exercises_target ON exercises(target_muscle);
-- Cursor pagination (name ASC is default sort)
CREATE INDEX idx_exercises_name_id ON exercises(name, id);

-- Populate search_vector on insert/update
ALTER TABLE exercises ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('english', name)) STORED;
```

### Table: `workout_templates`
```sql
CREATE TABLE workout_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_templates_user ON workout_templates(user_id);
CREATE INDEX idx_templates_user_created ON workout_templates(user_id, created_at DESC);
```

### Table: `template_slots`
```sql
CREATE TABLE template_slots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id     UUID NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    exercise_id     VARCHAR(10) NOT NULL REFERENCES exercises(id),
    position        INT NOT NULL,               -- ordering
    target_sets     INT,
    target_reps     INT,
    target_weight   DECIMAL(6,2),               -- kg
    rest_seconds    INT,
    UNIQUE(template_id, position)
);
CREATE INDEX idx_template_slots_template ON template_slots(template_id, position);
```

### Table: `workout_sessions`
```sql
CREATE TABLE workout_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id UUID REFERENCES workout_templates(id) ON DELETE SET NULL,  -- nullable (ad-hoc)
    started_at  TIMESTAMPTZ NOT NULL,
    ended_at    TIMESTAMPTZ,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user_started ON workout_sessions(user_id, started_at DESC);
CREATE INDEX idx_sessions_user ON workout_sessions(user_id);
```

### Table: `session_exercises`
```sql
CREATE TABLE session_exercises (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    exercise_id VARCHAR(10) NOT NULL REFERENCES exercises(id),
    position    INT NOT NULL,
    UNIQUE(session_id, position)
);
CREATE INDEX idx_session_exercises_session ON session_exercises(session_id, position);
CREATE INDEX idx_session_exercises_exercise ON session_exercises(exercise_id);
```

### Table: `session_sets`
```sql
CREATE TABLE session_sets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exercise_id     UUID NOT NULL REFERENCES session_exercises(id) ON DELETE CASCADE,
    position        INT NOT NULL,               -- set order
    reps            INT NOT NULL CHECK (reps > 0),
    weight          DECIMAL(6,2) CHECK (weight >= 0),
    duration_seconds INT CHECK (duration_seconds >= 0),
    UNIQUE(exercise_id, position)
);
CREATE INDEX idx_session_sets_exercise ON session_sets(exercise_id, position);
```

### Table: `personal_records`
```sql
CREATE TABLE personal_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id VARCHAR(10) NOT NULL REFERENCES exercises(id),
    max_weight  DECIMAL(6,2) NOT NULL DEFAULT 0,
    max_reps    INT NOT NULL DEFAULT 0,
    max_volume  DECIMAL(10,2) NOT NULL DEFAULT 0,  -- SUM(reps * weight) across sets
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, exercise_id)
);
CREATE INDEX idx_pr_user ON personal_records(user_id);
CREATE INDEX idx_pr_user_exercise ON personal_records(user_id, exercise_id);
```

## Domain Entities

```go
// internal/domain/user.go
type User struct {
    ID        uuid.UUID
    Email     string
    Password  string  // bcrypt hash (never serialized to JSON)
    CreatedAt time.Time
    UpdatedAt time.Time
}

// internal/domain/exercise.go
type Exercise struct {
    ID               string   // dataset ID
    Name             string
    Category         string
    BodyPart         string
    Equipment        string
    TargetMuscle     string
    MuscleGroup      string
    SecondaryMuscles []string
    Instructions     string
    GIFUrl           string
    ThumbnailURL     string
}

// internal/domain/template.go
type WorkoutTemplate struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Name        string
    Description string
    Slots       []TemplateSlot
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type TemplateSlot struct {
    ID           uuid.UUID
    TemplateID   uuid.UUID
    ExerciseID   string
    Position     int
    TargetSets   *int
    TargetReps   *int
    TargetWeight *float64
    RestSeconds  *int
}

// internal/domain/session.go
type WorkoutSession struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    TemplateID *uuid.UUID  // nil = ad-hoc
    StartedAt  time.Time
    EndedAt    *time.Time
    Notes      string
    Exercises  []SessionExercise
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type SessionExercise struct {
    ID         uuid.UUID
    SessionID  uuid.UUID
    ExerciseID string
    Position   int
    Sets       []SessionSet
}

type SessionSet struct {
    ID              uuid.UUID
    ExerciseID      uuid.UUID  // FK to session_exercises
    Position        int
    Reps            int
    Weight          *float64
    DurationSeconds *int
}

// internal/domain/personal_record.go
type PersonalRecord struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    ExerciseID string
    MaxWeight  float64
    MaxReps    int
    MaxVolume  float64
    UpdatedAt  time.Time
}
```

## Repository Interfaces

```go
// internal/repository/interfaces.go

type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type RefreshTokenRepository interface {
    Create(ctx context.Context, token *domain.RefreshToken) error
    FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
    Revoke(ctx context.Context, id uuid.UUID) error
    RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type ExerciseRepository interface {
    List(ctx context.Context, filter ExerciseFilter) (*cursor.Page[domain.Exercise], error)
    FindByID(ctx context.Context, id string) (*domain.Exercise, error)
    Exists(ctx context.Context, ids []string) (bool, error)
    BulkUpsert(ctx context.Context, exercises []domain.Exercise) error
}

type ExerciseFilter struct {
    Search      string
    Category    string
    BodyPart    string
    Equipment   string
    TargetMuscle string
    Cursor      string  // encoded cursor (name, id)
    Limit       int
}

type TemplateRepository interface {
    Create(ctx context.Context, tmpl *domain.WorkoutTemplate) error
    FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.WorkoutTemplate, error)
    List(ctx context.Context, userID uuid.UUID, page cursor.PageRequest) (*cursor.Page[domain.WorkoutTemplate], error)
    Update(ctx context.Context, tmpl *domain.WorkoutTemplate) error
    Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type SessionRepository interface {
    Create(ctx context.Context, session *domain.WorkoutSession) error
    FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.WorkoutSession, error)
    List(ctx context.Context, filter SessionFilter) (*cursor.Page[domain.WorkoutSession], error)
    Update(ctx context.Context, session *domain.WorkoutSession) error
    Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type SessionFilter struct {
    UserID uuid.UUID
    From   *time.Time
    To     *time.Time
    Cursor string
    Limit  int
}

type PersonalRecordRepository interface {
    Upsert(ctx context.Context, pr *domain.PersonalRecord) error
    FindByUserAndExercise(ctx context.Context, userID uuid.UUID, exerciseID string) (*domain.PersonalRecord, error)
    ListByUser(ctx context.Context, userID uuid.UUID, exerciseID *string, page cursor.PageRequest) (*cursor.Page[domain.PersonalRecord], error)
    RecalculateFromSessions(ctx context.Context, userID uuid.UUID, exerciseID string) (*domain.PersonalRecord, error)
}

type ProgressRepository interface {
    ExerciseHistory(ctx context.Context, userID uuid.UUID, exerciseID string, page cursor.PageRequest) (*cursor.Page[ExerciseHistoryEntry], error)
    Summary(ctx context.Context, userID uuid.UUID) (*SummaryStats, error)
}
```

## Cursor Pagination Pattern

```go
// pkg/cursor/cursor.go

// PageRequest is the input: opaque cursor + limit.
type PageRequest struct {
    Cursor string // base64-encoded cursor, empty = first page
    Limit  int    // default 20, max 100
}

// Page[T] is the output: items + next_cursor.
type Page[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"next_cursor,omitempty"` // empty = last page
    HasMore    bool   `json:"has_more"`
}

// Cursor encoding for exercises: base64(name + "\x00" + id)
// Cursor encoding for sessions: base64(started_at + "\x00" + id)
// Cursor encoding for templates: base64(created_at + "\x00" + id)
```

**SQL pattern (keyset):**
```sql
-- Exercises (name ASC, id ASC tiebreaker)
SELECT * FROM exercises
WHERE (name, id) > ($cursor_name, $cursor_id)
  AND category = $category   -- optional filter
  AND search_vector @@ plainto_tsquery('english', $search)  -- optional search
ORDER BY name, id
LIMIT $limit + 1;  -- fetch one extra to determine has_more

-- Sessions (started_at DESC, id DESC tiebreaker)
SELECT * FROM workout_sessions
WHERE user_id = $user_id
  AND (started_at, id) < ($cursor_started_at, $cursor_id)
ORDER BY started_at DESC, id DESC
LIMIT $limit + 1;
```

## HTTP API Design

### Auth Endpoints (public)
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/auth/register` | `auth.Register` | Create user |
| POST | `/api/v1/auth/login` | `auth.Login` | Get tokens |
| POST | `/api/v1/auth/refresh` | `auth.Refresh` | Rotate tokens |
| POST | `/api/v1/auth/logout` | `auth.Logout` | Revoke refresh token |

### Exercise Endpoints (public for read)
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/exercises` | `exercise.List` | List/search/filter with cursor |
| GET | `/api/v1/exercises/{id}` | `exercise.GetByID` | Full exercise details |

### Template Endpoints (auth required)
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/templates` | `template.Create` | Create template |
| GET | `/api/v1/templates` | `template.List` | List user's templates |
| GET | `/api/v1/templates/{id}` | `template.GetByID` | Get template with slots |
| PUT | `/api/v1/templates/{id}` | `template.Update` | Replace template |
| DELETE | `/api/v1/templates/{id}` | `template.Delete` | Hard delete |

### Session Endpoints (auth required)
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/sessions` | `session.Create` | Log session (triggers PR) |
| GET | `/api/v1/sessions` | `session.List` | List with cursor + date filter |
| GET | `/api/v1/sessions/{id}` | `session.GetByID` | Full session details |
| PUT | `/api/v1/sessions/{id}` | `session.Update` | Update (triggers PR) |
| DELETE | `/api/v1/sessions/{id}` | `session.Delete` | Delete (triggers PR recalc) |

### Progress Endpoints (auth required)
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/progress/records` | `progress.ListRecords` | PRs with cursor |
| GET | `/api/v1/progress/exercises/{exercise_id}/history` | `progress.ExerciseHistory` | Sets over time |
| GET | `/api/v1/progress/summary` | `progress.Summary` | Aggregate stats |

### Media Endpoints (public)
| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/media/gifs/{filename}` | static | Serve GIF files |
| GET | `/media/thumbnails/{filename}` | static | Serve thumbnail files |

### Error Response Format
```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid input",
        "details": [
            {"field": "email", "message": "invalid email format"},
            {"field": "password", "message": "must be at least 8 characters"}
        ]
    }
}
```

Error codes: `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `INTERNAL_ERROR`.

### Middleware Chain
```
Request → Recovery → Logger → [Auth (protected routes only)] → Handler
```

## Authentication Flow

```
Register:  email+password → bcrypt hash → store user → return 201
Login:     email+password → verify bcrypt → generate access JWT (15m) + refresh token (7d)
           → store refresh token hash in DB → return {access_token, refresh_token}
Refresh:   refresh_token → lookup hash in DB → check not revoked/expired
           → revoke old token → generate new pair → return {access_token, refresh_token}
Logout:    refresh_token → lookup hash → revoke (set revoked_at) → return 200
Protected: Authorization: Bearer <access_token> → middleware validates JWT
           → extract user_id → inject into context → handler reads from context
```

**JWT claims:** `{ sub: user_id, exp, iat, type: "access" }` — minimal, no sensitive data.

## Exercise Seeder Design

```
cmd/seed/main.go:
  1. Load config (DB connection, dataset path)
  2. Read exercises.json from exercises-dataset/data/
  3. Map JSON → domain.Exercise (extract English instructions, build media URLs)
  4. Copy/symlink images → media/thumbnails/{id}.jpg
  5. Bulk upsert into exercises table (ON CONFLICT DO UPDATE)
  6. Log count: "Seeded 1,324 exercises"
```

**Media URL mapping:**
- Dataset `image: "images/0001-2gPfomN.jpg"` → stored as `thumbnail_url: "/media/thumbnails/0001.jpg"`
- Dataset `gif_url: "videos/0001-2gPfomN.gif"` → stored as `gif_url: "/media/gifs/0001.gif"`
- Seeder copies files from dataset to `media/` directory with simplified names.

## Configuration

```go
// configs/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Media    MediaConfig
}

type ServerConfig struct {
    Port string `env:"PORT" envDefault:"8080"`
    Host string `env:"HOST" envDefault:"0.0.0.0"`
}

type DatabaseConfig struct {
    URL             string `env:"DATABASE_URL,required"`
    MaxOpenConns    int    `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
    MaxIdleConns    int    `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
    ConnMaxLifetime string `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
}

type JWTConfig struct {
    AccessSecret      string `env:"JWT_ACCESS_SECRET,required"`
    RefreshSecret     string `env:"JWT_REFRESH_SECRET,required"`
    AccessExpiry      time.Duration `env:"JWT_ACCESS_EXPIRY" envDefault:"15m"`
    RefreshExpiry     time.Duration `env:"JWT_REFRESH_EXPIRY" envDefault:"168h"` // 7d
}

type MediaConfig struct {
    RootDir      string `env:"MEDIA_ROOT_DIR" envDefault:"./media"`
    GIFsDir      string `env:"MEDIA_GIFS_DIR" envDefault:"gifs"`
    ThumbnailsDir string `env:"MEDIA_THUMBNAILS_DIR" envDefault:"thumbnails"`
    DatasetPath  string `env:"DATASET_PATH"` // path to exercises-dataset
}
```

## Docker Setup

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: gym-tracker
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

```dockerfile
# Dockerfile (multi-stage)
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/gym-tracker ./cmd/api

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/gym-tracker /bin/gym-tracker
COPY media/ /app/media/
EXPOSE 8080
CMD ["/bin/gym-tracker"]
```

## Testing Strategy

| Layer | What | Approach | Tools |
|-------|------|----------|-------|
| **Unit** (usecase) | Business logic, PR calculation, validation | Mock repository interfaces with `testify/mock` or manual fakes | `go test`, `testify` |
| **Integration** (repository) | SQL queries, transactions, cursor pagination | Test DB (docker postgres), real queries, truncate between tests | `testutil/db.go` helper, `database/sql` |
| **Handler** | HTTP request/response, status codes, middleware | `httptest.NewRecorder` + real usecases with mock repos | `net/http/httptest` |
| **E2E** | Full stack (future) | Not in MVP scope | — |

**Test data factories:**
```go
// internal/testutil/factories.go
func NewUser(overrides ...func(*domain.User)) *domain.User { ... }
func NewExercise(overrides ...func(*domain.Exercise)) *domain.Exercise { ... }
func NewTemplate(userID uuid.UUID, overrides ...func(*domain.WorkoutTemplate)) *domain.WorkoutTemplate { ... }
func NewSession(userID uuid.UUID, overrides ...func(*domain.WorkoutSession)) *domain.WorkoutSession { ... }
```

**TDD enforcement:** `strict_tdd: true` in config. RED → GREEN → REFACTOR for each unit. Run `go test -race -cover ./...` after every change.

## Error Handling

```go
// internal/domain/errors.go
type AppError struct {
    Code    string            // "NOT_FOUND", "CONFLICT", etc.
    Message string            // human-readable
    Status  int               // HTTP status code
    Details []FieldError      // validation details
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

// Sentinel errors
var (
    ErrNotFound     = &AppError{Code: "NOT_FOUND", Message: "resource not found", Status: 404}
    ErrConflict     = &AppError{Code: "CONFLICT", Message: "resource already exists", Status: 409}
    ErrUnauthorized = &AppError{Code: "UNAUTHORIZED", Message: "invalid credentials", Status: 401}
    ErrForbidden    = &AppError{Code: "FORBIDDEN", Message: "access denied", Status: 403}
)

// Error propagation:
// Repository → returns domain.ErrNotFound or wrapped SQL errors
// UseCase    → returns domain.AppError (may wrap repo errors)
// Handler    → checks type, writes JSON error response with correct status
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `cmd/api/main.go` | Modify | Full app bootstrap, DI, router |
| `cmd/seed/main.go` | Create | Exercise seeder CLI |
| `internal/domain/*.go` | Create | 6 entity files + errors |
| `internal/usecase/*.go` | Create | 5 usecase files + tests |
| `internal/repository/interfaces.go` | Create | All repository ports |
| `internal/repository/postgres/*.go` | Create | 6 repo implementations + tests |
| `internal/handler/*.go` | Create | 6 handler files + router + response helpers + tests |
| `internal/middlewares/*.go` | Create | Auth, logger, recovery |
| `pkg/cursor/` | Create | Cursor pagination utilities |
| `pkg/jwt/` | Create | JWT token utilities |
| `pkg/validator/` | Create | Input validation |
| `pkg/pagination/` | Create | Pagination types |
| `configs/config.go` | Create | Config loading |
| `migrations/*.sql` | Create | 6 migration pairs (up/down) |
| `docker-compose.yml` | Create | Dev PostgreSQL |
| `Dockerfile` | Create | Multi-stage Go build |
| `media/` | Create | Static media directory |

## Migration / Rollout

No migration required — greenfield project. Implementation order:

1. **Foundation**: Config, DB connection, migrations, cursor pkg, JWT pkg, error types
2. **Auth**: User entity → repo → usecase → handler → middleware
3. **Exercise catalog**: Entity → seeder → repo → usecase → handler → media serving
4. **Templates**: Entity → repo → usecase → handler
5. **Sessions**: Entity → repo → usecase (with PR trigger) → handler
6. **Progress**: PR entity → repo → usecase → handler
7. **Polish**: Docker, Makefile updates, integration tests

## Open Questions

- [ ] Should `session_sets.weight` allow NULL (bodyweight exercises where weight is implicit) or default to 0?
- [ ] Do we need a `notes` field on `session_sets` (e.g., "felt heavy")? Deferred to post-MVP.
- [ ] Rate limiting on auth endpoints — defer to post-MVP or add simple middleware now?
