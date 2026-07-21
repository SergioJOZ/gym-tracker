# Workout Sessions Specification

## Purpose

Log workout instances with per-exercise sets (reps, weight, duration). Sessions can be created from a template or ad-hoc.

## Requirements

### Requirement: Create Session

The system MUST create a workout session with a start time, optional template reference, and a list of exercises with their completed sets.

#### Scenario: Create session from template

- GIVEN an authenticated user and a valid template ID
- WHEN the user submits POST /sessions with template_id, started_at, and exercises (each with exercise_id and sets containing reps, weight, duration)
- THEN the system returns 201 with the created session including its ID

#### Scenario: Create ad-hoc session

- GIVEN an authenticated user
- WHEN the user submits POST /sessions without template_id, with started_at and exercises
- THEN the system returns 201 with the created session (template_id is null)

#### Scenario: Create session with invalid exercise

- GIVEN an exercise_id not in the catalog
- WHEN the user submits POST /sessions
- THEN the system returns 400 with validation error

#### Scenario: Create session with empty exercises

- GIVEN a request with no exercises
- WHEN the user submits POST /sessions
- THEN the system returns 400 (a session MUST have at least one exercise)

### Requirement: List User Sessions

The system MUST return the authenticated user's sessions with cursor-based pagination, ordered by started_at descending.

#### Scenario: List sessions

- GIVEN an authenticated user with 5 sessions
- WHEN the user requests GET /sessions?limit=10
- THEN the system returns sessions ordered by started_at desc with a next_cursor

#### Scenario: Filter by date range

- GIVEN sessions on various dates
- WHEN the user requests GET /sessions?from=2026-01-01&to=2026-01-31
- THEN the system returns only sessions within that date range

### Requirement: Get Session by ID

The system MUST return a single session with full exercise and set details.

#### Scenario: Own session

- GIVEN an authenticated user who owns session "s-456"
- WHEN the user requests GET /sessions/s-456
- THEN the system returns 200 with session details including all exercises and their sets

#### Scenario: Another user's session

- GIVEN a session owned by a different user
- WHEN the user requests GET /sessions/{other-user-session-id}
- THEN the system returns 404

### Requirement: Update Session

The system MUST allow updating session details before completion.

#### Scenario: Update session exercises/sets

- GIVEN an authenticated user who owns session "s-456"
- WHEN the user submits PUT /sessions/s-456 with updated exercises and sets
- THEN the system returns 200 with the updated session

#### Scenario: Update triggers PR recalculation

- GIVEN a session update that changes set weights or reps
- WHEN the session is saved
- THEN the system recalculates personal records for affected exercises (see personal-records spec)

### Requirement: Delete Session

The system MUST allow deleting a session. Deletion MUST trigger PR recalculation.

#### Scenario: Delete own session

- GIVEN an authenticated user who owns session "s-456"
- WHEN the user submits DELETE /sessions/s-456
- THEN the system removes the session and returns 204
- AND personal records are recalculated for affected exercises

#### Scenario: Delete another user's session

- GIVEN a session owned by a different user
- WHEN the user submits DELETE /sessions/{other-user-session-id}
- THEN the system returns 404

## Constraints

- Sessions are per-user (user_id scoping)
- Each set MUST have at least reps (positive integer); weight and duration are optional
- Weight MUST be >= 0; duration MUST be >= 0 if provided
- Session deletion MUST trigger PR recalculation
- Maximum 20 exercises per session; maximum 10 sets per exercise (MVP limits)

## Dependencies

- auth (user authentication)
- exercise-catalog (exercise IDs must exist)
- personal-records (PR recalculation on create/update/delete)
