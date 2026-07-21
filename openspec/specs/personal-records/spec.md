# Personal Records Specification

## Purpose

Auto-calculate and store personal records per user per exercise: max_weight, max_reps, max_volume. PRs are updated automatically when sessions are created, updated, or deleted.

## Requirements

### Requirement: Auto-calculate PRs on Session Save

The system MUST recalculate personal records for each exercise in a session whenever the session is created or updated.

#### Scenario: New max weight

- GIVEN a user whose PR for "bench press" is 80kg
- WHEN the user logs a session with bench press at 85kg
- THEN the system updates the PR max_weight to 85kg

#### Scenario: New max reps

- GIVEN a user whose PR for "pull-ups" is 10 reps (at any weight)
- WHEN the user logs a session with 12 reps of pull-ups
- THEN the system updates the PR max_reps to 12

#### Scenario: New max volume

- GIVEN a user whose PR for "squat" max_volume is 500kg (sets * reps * weight)
- WHEN the user logs a session with squat: 3 sets x 5 reps x 40kg = 600kg total volume
- THEN the system updates the PR max_volume to 600kg

#### Scenario: No new PR

- GIVEN a user whose PRs are max_weight=100kg, max_reps=15, max_volume=800kg
- WHEN the user logs a session with values below all PRs
- THEN the PR records remain unchanged

### Requirement: Recalculate PRs on Session Delete

The system MUST recalculate PRs from remaining sessions when a session is deleted.

#### Scenario: Delete session containing a PR

- GIVEN a user whose max_weight PR for "bench press" (100kg) came from session "s-1"
- WHEN session "s-1" is deleted
- THEN the system recalculates max_weight from remaining sessions
- AND the PR reflects the new max from remaining data

#### Scenario: Delete session not containing any PR

- GIVEN a session with values below all PRs
- WHEN the session is deleted
- THEN PR records remain unchanged

### Requirement: PR Record Structure

Each PR record MUST contain: user_id, exercise_id, max_weight, max_reps, max_volume, updated_at.

#### Scenario: PR record created on first session for exercise

- GIVEN a user logging their first session for "deadlift"
- WHEN the session is saved
- THEN a PR record is created with max_weight, max_reps, and max_volume from that session

#### Scenario: PR record updated on subsequent session

- GIVEN an existing PR record for user + exercise
- WHEN a new session exceeds any PR value
- THEN the PR record is updated (upsert) with new max values and updated_at

## Constraints

- PR calculation MUST be atomic (use DB transactions with row-level locks)
- max_volume = SUM(reps * weight) across all sets for that exercise in a session
- max_reps = highest single-set rep count for that exercise
- max_weight = highest single-set weight for that exercise
- PR updates MUST happen in the same transaction as the session save (consistency)
- No 1RM calculation in MVP

## Dependencies

- workout-sessions (PRs are derived from session data)
- exercise-catalog (exercise_id references)
