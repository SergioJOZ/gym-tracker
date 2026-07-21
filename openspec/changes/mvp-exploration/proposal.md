# Proposal: Gym-Tracker MVP

## Intent

Build an individual fitness tracking backend (like Hevy without social features) that lets users browse a curated exercise catalog, create workout templates, log workout sessions with sets/reps/weight, and track personal records over time.

## Scope

### In Scope
- JWT auth (register, login, refresh, logout) for individual users
- Exercise catalog: 1,324 exercises from dataset (read-only, seeded once)
- Exercise search/filter (by name, category, body part, equipment, target muscle)
- Workout templates: CRUD with ordered exercises and target sets/reps/weight/rest
- Workout sessions: log from template or ad-hoc, with per-exercise sets
- Personal records: max_weight, max_reps, max_volume (auto-updated on session save)
- Progress endpoints: PR list, per-exercise history, summary stats
- Media: GIFs and thumbnails served as static files from Go backend
- Cursor-based pagination for exercises and sessions

### Out of Scope
- Custom user-created exercises
- Multi-tenant (gym/coach/client model)
- Social features (sharing, following, leaderboards)
- 1RM calculation or advanced analytics
- Flutter frontend (API-only for MVP)
- CDN for media (static files from Go)
- Offline support / sync
- Exercise instructions in multiple languages (stored but not exposed in MVP API)

## Capabilities

### New Capabilities
- `auth`: JWT registration, login, token refresh, logout
- `exercise-catalog`: Read-only exercise browsing, search, and filtering from seeded dataset
- `workout-templates`: CRUD for reusable workout routines with exercise slots
- `workout-sessions`: Logging workout instances with per-exercise sets
- `personal-records`: Auto-calculated PRs (max_weight, max_reps, max_volume) per user per exercise
- `progress`: PR listing, exercise history, summary statistics
- `media-serving`: Static file serving for exercise GIFs and thumbnails

### Modified Capabilities
None (greenfield project)

## Approach

Monolithic Go binary with clean/hexagonal architecture. Modules: auth, exercise, workout, progress. PostgreSQL with cursor-based pagination. Exercise data seeded from `exercises-dataset` via CLI command. Media served via `http.FileServer`. Strict TDD: test first at each layer (domain → usecase → repository → handler).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/domain/` | New | All entities: User, Exercise, WorkoutTemplate, WorkoutSession, PersonalRecord |
| `internal/usecase/` | New | Business logic for auth, exercises, templates, sessions, PRs |
| `internal/handler/` | New | REST API handlers for all endpoints |
| `internal/repository/postgres/` | New | PostgreSQL implementations for all repositories |
| `internal/middlewares/` | New | JWT auth middleware |
| `cmd/api/main.go` | New | App bootstrap, DI wiring, router setup |
| `cmd/api/seed/main.go` | New | Exercise seeder CLI |
| `migrations/` | New | Schema migrations for all tables |
| `pkg/validator/` | New | Request validation utilities |
| `media/` | New | Static directory for GIFs and thumbnails |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| GIF bandwidth from Go backend | High | Serve only thumbnails from backend; GIFs via configurable URL (CDN-ready) |
| PR update race conditions | Medium | Use DB transactions with row-level locks on personal_records |
| Exercise search performance on 1,324 rows | Low | GIN index on name tsvector; verify with EXPLAIN ANALYZE |
| Multi-tenant migration later | Medium | All tables already have user_id; tenant_id can be added as nullable column |
| Cursor pagination complexity vs offset | Low | Well-documented keyset pattern; first/last support |

## Rollback Plan

MVP is greenfield — rollback means dropping the database and removing the binary. Each module is independently deployable behind feature flags if needed. Database migrations are forward-only for MVP; rollback requires `DROP TABLE` in reverse order.

## Dependencies

- `exercises-dataset` repo at `/home/ventuzzn/Documents/projects/exercises-dataset/` (1,324 exercises JSON + media)
- PostgreSQL 15+ (via Docker)
- Go 1.26.4

## Success Criteria

- [ ] User can register, login, and access protected endpoints via JWT
- [ ] Exercise catalog is seeded and searchable (by name, category, body part, equipment)
- [ ] User can create/edit/delete workout templates with exercises
- [ ] User can log a workout session with multiple exercises and sets
- [ ] Personal records update automatically when sessions are saved
- [ ] Progress endpoints return PR list, exercise history, and summary
- [ ] All tests pass: `go test -race -cover ./...` with ≥80% coverage
- [ ] API serves exercise thumbnails and GIFs as static files
