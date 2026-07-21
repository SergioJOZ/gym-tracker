# Tasks: Gym-Tracker MVP

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 5500–7000 (60+ files, 8 migrations, 7 capabilities) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 → PR 4 → PR 5 |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision resolved: chained PRs with stacked-to-main strategy
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation + Auth (config, DB, migrations, cursor, JWT, errors, auth CRUD, middleware) | PR 1 | base: main; all subsequent PRs depend on this |
| 2 | Exercise Catalog (entity, seeder CLI, repo, usecase, handler, media serving) | PR 2 | base: main (stacked) or PR 1 branch; standalone read-only module |
| 3 | Workout Templates (entity, repo, usecase, handler with cursor pagination) | PR 3 | depends on PR 1 (auth); independent of exercises |
| 4 | Sessions + PR Trigger (entity, repo, usecase with inline PR calc, handler) | PR 4 | depends on PR 1 + PR 3 (references exercises and templates) |
| 5 | Progress + Polish (PR entity, progress endpoints, router wiring, Docker, seed CLI) | PR 5 | depends on PR 4; integration tests and deployment |

## Phase 1: Foundation

- [x] 1.1 Create `configs/config.go` — Config struct with env loading (caarlos0/env), Server/Database/JWT/Media sections
- [x] 1.2 Create `docker-compose.yml` — PostgreSQL 16-alpine service, port 5432, pgdata volume
- [x] 1.3 Create `migrations/000001_create_users.{up,down}.sql` — users table + email index
- [ ] 1.4 Create `migrations/000002_create_exercises.{up,down}.sql` — exercises table, tsvector, GIN + filter indexes
- [ ] 1.5 Create `migrations/000003_create_templates.{up,down}.sql` — workout_templates + template_slots tables
- [ ] 1.6 Create `migrations/000004_create_sessions.{up,down}.sql` — workout_sessions + session_exercises + session_sets tables
- [ ] 1.7 Create `migrations/000005_create_personal_records.{up,down}.sql` — personal_records table + unique(user_id, exercise_id)
- [x] 1.8 Create `migrations/000006_create_refresh_tokens.{up,down}.sql` — refresh_tokens table + hash/user indexes
- [x] 1.9 Create `pkg/cursor/cursor.go` + `cursor_test.go` — Encode/Decode base64 cursors, PageRequest, Page[T] generic types
- [x] 1.10 Create `pkg/jwt/jwt.go` + `jwt_test.go` — GenerateAccess/RefreshToken, ValidateToken, Claims struct (sub, exp, iat, type)
- [x] 1.11 Create `internal/domain/errors.go` — AppError type, FieldError, sentinel errors (ErrNotFound, ErrConflict, ErrUnauthorized, ErrForbidden)
- [x] 1.12 Create `pkg/validator/validator.go` + `validator_test.go` — Validate struct helper using go-playground/validator, returns []FieldError

## Phase 2: Auth

- [x] 2.1 Create `internal/domain/user.go` — User entity + RefreshToken entity (ID, Email, Password/TokenHash, timestamps)
- [x] 2.2 Create `internal/repository/interfaces.go` — UserRepository + RefreshTokenRepository interfaces (ports)
- [x] 2.3 Create `internal/repository/postgres/testutil/db.go` — Test DB setup/teardown, truncate helpers, connection pool
- [x] 2.4 Create `internal/repository/postgres/user.go` + `user_test.go` — Create, FindByEmail, FindByID with integration tests
- [x] 2.5 Create `internal/repository/postgres/refresh_token.go` + `refresh_token_test.go` — Create, FindByHash, Revoke, RevokeAllForUser
- [x] 2.6 Create `internal/usecase/auth.go` + `auth_test.go` — Register (bcrypt>=10), Login (JWT+refresh), Refresh (rotate), Logout (revoke); mock repo tests
- [x] 2.7 Create `internal/handler/response.go` — JSON response helpers: respondJSON, respondError, AppError→JSON mapping
- [x] 2.8 Create `internal/handler/request.go` — DecodeJSONBody, parsePaginationParams (cursor, limit with defaults)
- [x] 2.9 Create `internal/handler/auth.go` + `auth_test.go` — POST register/login/refresh/logout handlers; httptest tests
- [x] 2.10 Create `internal/middlewares/auth.go` + `auth_test.go` — JWT extraction from Authorization header, user_id→context, 401 on invalid

## Phase 3: Exercise Catalog

- [ ] 3.1 Create `internal/domain/exercise.go` — Exercise entity (ID, Name, Category, BodyPart, Equipment, TargetMuscle, MuscleGroup, SecondaryMuscles, Instructions, GIFUrl, ThumbnailURL)
- [ ] 3.2 Add ExerciseRepository interface to `internal/repository/interfaces.go` — List(filter), FindByID, Exists, BulkUpsert + ExerciseFilter struct
- [ ] 3.3 Create `internal/repository/postgres/exercise.go` + `exercise_test.go` — Cursor keyset query, FTS search, filter AND logic, BulkUpsert
- [ ] 3.4 Create `internal/usecase/exercise.go` + `exercise_test.go` — List, GetByID with filter validation; mock repo tests
- [ ] 3.5 Create `internal/handler/exercise.go` + `exercise_test.go` — GET /exercises (list/search/filter), GET /exercises/{id}; httptest tests
- [ ] 3.6 Create `internal/handler/media.go` — http.FileServer for /media/gifs/ and /media/thumbnails/ with path traversal protection
- [ ] 3.7 Create `cmd/seed/main.go` — Load exercises.json from dataset, map→domain, copy media files, BulkUpsert, log count

## Phase 4: Workout Templates

- [ ] 4.1 Create `internal/domain/template.go` — WorkoutTemplate + TemplateSlot entities
- [ ] 4.2 Add TemplateRepository interface to `internal/repository/interfaces.go` — CRUD + List with cursor
- [ ] 4.3 Create `internal/repository/postgres/template.go` + `template_test.go` — CRUD with slots in tx, cursor pagination, user-scoped queries
- [ ] 4.4 Create `internal/usecase/template.go` + `template_test.go` — CRUD with validation (max 50 slots, exercise exists check)
- [ ] 4.5 Create `internal/handler/template.go` + `template_test.go` — POST/GET/PUT/DELETE /templates; httptest with mock usecase

## Phase 5: Workout Sessions + PR Trigger

- [ ] 5.1 Create `internal/domain/session.go` — WorkoutSession, SessionExercise, SessionSet entities
- [ ] 5.2 Create `internal/domain/personal_record.go` — PersonalRecord entity (UserID, ExerciseID, MaxWeight, MaxReps, MaxVolume)
- [ ] 5.3 Add SessionRepository + PersonalRecordRepository interfaces to `internal/repository/interfaces.go`
- [ ] 5.4 Create `internal/repository/postgres/session.go` + `session_test.go` — CRUD with nested exercises/sets in tx, cursor+date filter
- [ ] 5.5 Create `internal/repository/postgres/personal_record.go` + `personal_record_test.go` — Upsert (GREATEST), FindByUserAndExercise, RecalculateFromSessions
- [ ] 5.6 Create `internal/usecase/session.go` + `session_test.go` — CRUD with inline PR calc in tx (create/update: UPSERT GREATEST; delete: recompute from remaining)
- [ ] 5.7 Create `internal/handler/session.go` + `session_test.go` — POST/GET/PUT/DELETE /sessions; httptest with mock usecase

## Phase 6: Progress + Wiring + Polish

- [ ] 6.1 Add ProgressRepository interface to `internal/repository/interfaces.go` — ExerciseHistory, Summary
- [ ] 6.2 Create `internal/repository/postgres/progress.go` + `progress_test.go` — ExerciseHistory (chronological sets), Summary (aggregate stats)
- [ ] 6.3 Create `internal/usecase/progress.go` + `progress_test.go` — ListPRs, ExerciseHistory, Summary
- [ ] 6.4 Create `internal/handler/progress.go` + `progress_test.go` — GET /progress/records, /exercises/{id}/history, /summary
- [ ] 6.5 Create `internal/handler/router.go` — chi router setup, route registration, middleware chain (Recovery→Logger→[Auth]→Handler)
- [ ] 6.6 Create `internal/middlewares/logger.go` + `recovery.go` — Request logging middleware, panic recovery middleware
- [ ] 6.7 Update `cmd/api/main.go` — Full bootstrap: config→DB→repos→usecases→handlers→router→listen
- [ ] 6.8 Create `Dockerfile` — Multi-stage build (golang:1.26-alpine builder, alpine:3.19 runtime)
- [ ] 6.9 Create `internal/testutil/factories.go` — Test data factories: NewUser, NewExercise, NewTemplate, NewSession
- [ ] 6.10 Update `Makefile` — Add seed target, test-integration target with test DB
