## Exploration: Gym-Tracker MVP

### Current State
The project is a greenfield Go backend with clean/hexagonal architecture scaffolding. All internal directories (domain, handler, usecase, repository, middlewares) are empty. `main.go` is a placeholder. No dependencies in `go.mod` beyond the module declaration. PostgreSQL is configured via Makefile. The exercises-dataset (1,324 exercises with multilingual instructions, GIFs, and thumbnails) exists as a sibling repo at `/home/ventuzzn/Documents/projects/exercises-dataset/`.

### Affected Areas
- `internal/domain/` — All domain entities need to be created from scratch
- `internal/usecase/` — Business logic for auth, exercises, workouts, progress
- `internal/handler/` — HTTP handlers for REST API
- `internal/repository/` — PostgreSQL implementations
- `internal/middlewares/` — Auth middleware (JWT)
- `migrations/` — Database schema migrations
- `cmd/api/main.go` — Application bootstrap, DI wiring
- `pkg/validator/` — Request validation utilities

### Domain Entities & Relationships

```
User (1) ──── (N) WorkoutTemplate
  │                    │
  │                    └── (N) WorkoutTemplateExercise ──── Exercise (catalog)
  │
  └──── (N) WorkoutSession
           │
           └── (N) WorkoutSessionExercise ──── Exercise (catalog)
                    │
                    └── (N) ExerciseSet

User (1) ──── (N) PersonalRecord ──── Exercise (catalog)
```

#### Entity Definitions

1. **User** — Individual app user (no multi-tenant for MVP)
   - `ID` (UUID), `Email` (unique), `PasswordHash`, `Name`, `CreatedAt`, `UpdatedAt`

2. **Exercise** — Catalog entity from exercises-dataset (read-only for MVP)
   - `ID` (string "0001"-"1324"), `Name`, `Category`, `BodyPart`, `Equipment`
   - `Instructions` (JSONB: {en, es, it, tr, ru, zh})
   - `InstructionSteps` (JSONB: {en: [...], es: [...], ...})
   - `MuscleGroup`, `SecondaryMuscles` ([]string), `Target`
   - `ImageURL`, `GifURL`, `MediaID`, `Attribution`, `CreatedAt`

3. **WorkoutTemplate** — User-created reusable workout routine
   - `ID` (UUID), `UserID` (FK), `Name`, `Description`, `CreatedAt`, `UpdatedAt`

4. **WorkoutTemplateExercise** — Exercise slot within a template
   - `ID` (UUID), `WorkoutTemplateID` (FK), `ExerciseID` (FK)
   - `Order`, `TargetSets`, `TargetReps`, `TargetWeight`, `RestSeconds`

5. **WorkoutSession** — A logged workout instance
   - `ID` (UUID), `UserID` (FK), `WorkoutTemplateID` (FK, nullable for ad-hoc)
   - `Name`, `Notes`, `StartedAt`, `CompletedAt`, `CreatedAt`

6. **WorkoutSessionExercise** — Exercise within a logged session
   - `ID` (UUID), `WorkoutSessionID` (FK), `ExerciseID` (FK)
   - `Order`, `Notes`

7. **ExerciseSet** — Individual set logged within a session exercise
   - `ID` (UUID), `WorkoutSessionExerciseID` (FK)
   - `SetNumber`, `Reps`, `Weight` (decimal), `DurationSeconds`, `RestSeconds`, `CompletedAt`

8. **PersonalRecord** — Materialized PR per user per exercise per type
   - `ID` (UUID), `UserID` (FK), `ExerciseID` (FK)
   - `RecordType` (max_weight | max_reps | max_volume), `Value`, `ExerciseSetID` (FK), `AchievedAt`
   - Unique constraint: (user_id, exercise_id, record_type)

### User Flows

#### Flow 1: Log a Workout Session
```
User → Start Workout → Choose Template OR Ad-hoc
  → For each exercise:
    → View exercise details (name, GIF, instructions)
    → Log sets: reps, weight, rest time
    → Mark set complete
  → Finish workout
  → System: save session, calculate & update PRs
```

#### Flow 2: Create a Workout Template
```
User → Templates → Create New
  → Set name, description
  → Search exercise catalog → Add exercises
  → Order exercises, set targets (sets/reps/weight/rest)
  → Save template
```

#### Flow 3: View Progress
```
User → Progress/Dashboard
  → Recent workout sessions timeline
  → Per-exercise history (all sessions where exercise was used)
  → Personal records list (heaviest weight, most reps, max volume)
  → Charts: weight progression, weekly volume, frequency
```

### Proposed Database Schema

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Exercises (seeded from exercises-dataset, read-only catalog)
CREATE TABLE exercises (
    id VARCHAR(4) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    body_part VARCHAR(100) NOT NULL,
    equipment VARCHAR(100) NOT NULL,
    instructions JSONB NOT NULL DEFAULT '{}',
    instruction_steps JSONB NOT NULL DEFAULT '{}',
    muscle_group VARCHAR(100),
    secondary_muscles TEXT[] DEFAULT '{}',
    target VARCHAR(100) NOT NULL,
    image_url VARCHAR(500),
    gif_url VARCHAR(500),
    media_id VARCHAR(50),
    attribution TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exercises_category ON exercises(category);
CREATE INDEX idx_exercises_equipment ON exercises(equipment);
CREATE INDEX idx_exercises_target ON exercises(target);
CREATE INDEX idx_exercises_body_part ON exercises(body_part);
CREATE INDEX idx_exercises_name ON exercises USING gin (to_tsvector('english', name));

-- Workout Templates
CREATE TABLE workout_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_templates_user ON workout_templates(user_id);

-- Template Exercises
CREATE TABLE workout_template_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_template_id UUID NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    exercise_id VARCHAR(4) NOT NULL REFERENCES exercises(id),
    exercise_order INT NOT NULL,
    target_sets INT,
    target_reps INT,
    target_weight DECIMAL(6,2),
    rest_seconds INT,
    UNIQUE(workout_template_id, exercise_order)
);

-- Workout Sessions
CREATE TABLE workout_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workout_template_id UUID REFERENCES workout_templates(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    notes TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workout_sessions_user ON workout_sessions(user_id);
CREATE INDEX idx_workout_sessions_started ON workout_sessions(started_at DESC);

-- Session Exercises
CREATE TABLE workout_session_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_session_id UUID NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    exercise_id VARCHAR(4) NOT NULL REFERENCES exercises(id),
    exercise_order INT NOT NULL,
    notes TEXT
);

-- Exercise Sets
CREATE TABLE exercise_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_session_exercise_id UUID NOT NULL REFERENCES workout_session_exercises(id) ON DELETE CASCADE,
    set_number INT NOT NULL,
    reps INT NOT NULL,
    weight DECIMAL(6,2),
    duration_seconds INT,
    rest_seconds INT,
    completed_at TIMESTAMPTZ,
    UNIQUE(workout_session_exercise_id, set_number)
);

-- Personal Records (materialized, auto-updated)
CREATE TABLE personal_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id VARCHAR(4) NOT NULL REFERENCES exercises(id),
    record_type VARCHAR(50) NOT NULL,
    value DECIMAL(10,2) NOT NULL,
    exercise_set_id UUID REFERENCES exercise_sets(id) ON DELETE SET NULL,
    achieved_at TIMESTAMPTZ NOT NULL,
    UNIQUE(user_id, exercise_id, record_type)
);

CREATE INDEX idx_personal_records_user_exercise ON personal_records(user_id, exercise_id);
```

### REST API Endpoints

#### Auth (`/api/v1/auth`)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/register` | Create account |
| POST | `/login` | Authenticate, return JWT |
| POST | `/refresh` | Refresh access token |
| POST | `/logout` | Invalidate refresh token |

#### Exercises (`/api/v1/exercises`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | List exercises (search, filter, paginate) |
| GET | `/:id` | Get exercise detail |
| GET | `/categories` | List distinct categories |
| GET | `/equipment` | List distinct equipment types |
| GET | `/targets` | List distinct target muscles |

#### Workout Templates (`/api/v1/workouts/templates`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | List user's templates |
| POST | `/` | Create template |
| GET | `/:id` | Get template with exercises |
| PUT | `/:id` | Update template |
| DELETE | `/:id` | Delete template |

#### Workout Sessions (`/api/v1/workouts/sessions`)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/` | Start/log a session |
| GET | `/` | List user's sessions (paginated) |
| GET | `/:id` | Get session with exercises & sets |
| PUT | `/:id` | Update/complete session |
| DELETE | `/:id` | Delete session |

#### Progress (`/api/v1/progress`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/records` | List user's PRs |
| GET | `/exercises/:id/history` | Exercise history across sessions |
| GET | `/summary` | Overview stats (total workouts, streak, etc.) |

### Exercise Module Architecture (Clean/Hexagonal)

```
internal/
  domain/
    exercise.go              -- Exercise entity + value objects
    exercise_repository.go   -- Repository interface (port)
  
  usecase/
    exercise/
      list_exercises.go      -- Search/filter/paginate catalog
      get_exercise.go        -- Get single exercise by ID
      seed_exercises.go      -- Import exercises.json to DB
  
  handler/
    exercise_handler.go      -- HTTP handlers (adapter)
  
  repository/
    postgres/
      exercise_repository.go -- PostgreSQL implementation (adapter)

cmd/
  api/
    seed/
      main.go               -- CLI seeder command
```

#### Seeder Strategy
- **Recommendation**: One-time CLI command (`cmd/api/seed/main.go`) that reads `exercises.json` and bulk-inserts into PostgreSQL
- Run once after initial migration: `go run ./cmd/api/seed`
- Idempotent: skip exercises that already exist (ON CONFLICT DO NOTHING)
- Media URLs stored as relative paths; served via configurable base URL
- Future: add `--force` flag to re-seed if dataset updates

#### Media Serving Strategy
| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| **Static from backend** | Simple, no external deps | Bandwidth cost, slow for GIFs | Low |
| **CDN URLs in DB** | Fast, scalable | Requires CDN setup | Medium |
| **Hybrid (recommended for MVP)** | Thumbnails from backend, GIFs via CDN path | Slightly more config | Low |

**MVP Recommendation**: Store relative paths in DB (`/media/images/0001.jpg`, `/media/videos/0001.gif`). Serve from Go static file handler initially. Later migrate to CDN by changing base URL config.

### Authentication Architecture
- JWT-based (access token + refresh token)
- Access token: short-lived (15min), in-memory on client
- Refresh token: long-lived (7d), httpOnly cookie or stored securely
- Password hashing: bcrypt
- Middleware extracts user ID from JWT for protected routes
- All user-scoped resources filtered by `user_id` from token

```
internal/
  domain/
    user.go                  -- User entity
    user_repository.go       -- Repository interface
    auth.go                  -- Auth value objects (LoginInput, TokenPair)
  
  usecase/
    auth/
      register.go
      login.go
      refresh_token.go
  
  handler/
    auth_handler.go
  
  repository/
    postgres/
      user_repository.go
  
  middlewares/
    auth.go                  -- JWT validation middleware
```

### Testing Strategy

| Layer | What to Test | Tooling |
|-------|-------------|---------|
| **Domain** | Entity validation, business rules | Standard `testing` |
| **Use Cases** | Business logic with mock repositories | `testing` + interfaces + manual mocks |
| **Repositories** | SQL correctness, constraints | Integration tests with test DB (testcontainers or docker-compose) |
| **Handlers** | HTTP request/response, status codes | `httptest` + mock use cases |
| **E2E** | Full flow (future) | Not for MVP |

**TDD approach** (per `openspec/config.yaml` strict_tdd: true):
1. Write failing test for use case
2. Implement domain + use case
3. Write failing test for repository (integration)
4. Implement repository
5. Write failing test for handler
6. Implement handler
7. Refactor

**Test commands**: `go test -race -cover ./...`

### Approaches

1. **Monolithic modules (recommended)** — All modules in single Go binary, clean architecture layers
   - Pros: Simple deployment, easy to develop, clear boundaries for future extraction
   - Cons: Single deployable unit
   - Effort: Medium

2. **Microservices per module** — Separate services for auth, exercises, workouts
   - Pros: Independent scaling, team autonomy
   - Cons: Over-engineering for MVP, operational complexity, network calls
   - Effort: High

3. **Modular monolith with event bus** — Modules communicate via internal events
   - Pros: Clean decoupling, easier to extract later
   - Cons: Added complexity for MVP, event sourcing overhead
   - Effort: High

### Recommendation
**Approach 1: Monolithic with clean architecture.** This is the right choice for MVP. The clean/hexagonal architecture already provides clear boundaries. Each module (auth, exercise, workout, progress) lives in its own package under `internal/`. The domain layer has zero external dependencies. Repository interfaces allow swapping implementations. This can be split into microservices later if needed.

### Risks
- **Exercise dataset size**: 1,324 exercises with JSONB instructions — need to verify query performance with proper indexing
- **Media storage**: GIFs are ~180x180 but 1,324 of them still need hosting strategy
- **PR calculation**: Real-time PR updates on every set completion need careful transaction handling
- **Multi-tenant migration**: MVP schema has `user_id` everywhere — adding tenant_id later requires migration planning
- **Flutter integration**: API design should consider mobile-first patterns (pagination, offline support)

### Open Questions / Decisions Needed
1. **Exercise catalog mutability**: Should users be able to create custom exercises, or is the catalog read-only for MVP?
2. **Workout session status**: Should sessions have explicit states (active, paused, completed)?
3. **PR types for MVP**: Which PR types to support initially? (max_weight, max_reps, max_volume — or more?)
4. **Media hosting**: Static files from Go backend for MVP, or set up CDN from the start?
5. **Pagination strategy**: Offset-based or cursor-based for exercise list and session history?
6. **Exercise search**: Full-text search on name only, or also search instructions/muscles?

### Ready for Proposal
**Yes.** The exploration provides enough context for the orchestrator to proceed to `sdd-propose`. The user should confirm:
- Exercise catalog is read-only for MVP (no custom exercises)
- Media served from backend static files initially
- PR types: max_weight, max_reps, max_volume
- Offset-based pagination for MVP simplicity
