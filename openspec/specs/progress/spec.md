# Progress Specification

## Purpose

Endpoints for users to view their personal records, exercise history, and summary statistics over time.

## Requirements

### Requirement: List Personal Records

The system MUST return all PR records for the authenticated user with cursor-based pagination.

#### Scenario: List all PRs

- GIVEN an authenticated user with PRs for 5 exercises
- WHEN the user requests GET /progress/records
- THEN the system returns PR records (exercise_id, max_weight, max_reps, max_volume, updated_at)

#### Scenario: Filter PRs by exercise

- GIVEN PRs for multiple exercises
- WHEN the user requests GET /progress/records?exercise_id=abc-123
- THEN the system returns only the PR for that exercise

#### Scenario: No PRs yet

- GIVEN an authenticated user with no sessions
- WHEN the user requests GET /progress/records
- THEN the system returns 200 with an empty list

### Requirement: Exercise History

The system MUST return session history for a specific exercise, showing all sets over time.

#### Scenario: Exercise history with data

- GIVEN an authenticated user with 3 sessions containing "bench press"
- WHEN the user requests GET /progress/exercises/{exercise_id}/history
- THEN the system returns chronological list of sets (date, reps, weight, volume)

#### Scenario: Exercise history with no data

- GIVEN an authenticated user with no sessions for exercise "xyz"
- WHEN the user requests GET /progress/exercises/xyz/history
- THEN the system returns 200 with an empty list

### Requirement: Summary Statistics

The system MUST return aggregate statistics for the authenticated user.

#### Scenario: Summary with data

- GIVEN an authenticated user with 10 sessions across 20 exercises
- WHEN the user requests GET /progress/summary
- THEN the system returns 200 with total_sessions, total_workouts (unique exercises), total_volume (all-time), and active_days (unique dates)

#### Scenario: Summary with no data

- GIVEN an authenticated user with no sessions
- WHEN the user requests GET /progress/summary
- THEN the system returns 200 with all stats at 0

## Constraints

- All progress endpoints are per-user (user_id scoping)
- Summary statistics SHOULD be cached or computed efficiently (avoid full table scans on large datasets)
- Exercise history MUST be ordered chronologically (oldest first)
- Cursor-based pagination for PR list and exercise history

## Dependencies

- auth (user authentication)
- personal-records (PR data source)
- workout-sessions (session data source)
- exercise-catalog (exercise details for display)
