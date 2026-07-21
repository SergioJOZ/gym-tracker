# Workout Templates Specification

## Purpose

CRUD for reusable workout routines. Each template contains an ordered list of exercise slots with target sets, reps, weight, and rest time.

## Requirements

### Requirement: Create Template

The system MUST create a workout template with a name, optional description, and an ordered list of exercise slots. Each slot references an exercise from the catalog.

#### Scenario: Create template with exercises

- GIVEN an authenticated user and valid exercise IDs
- WHEN the user submits POST /templates with name, exercises (ordered list with exercise_id, target_sets, target_reps, target_weight, rest_seconds)
- THEN the system returns 201 with the created template including its ID and exercise slots

#### Scenario: Create template without exercises

- GIVEN an authenticated user
- WHEN the user submits POST /templates with name only (empty exercises list)
- THEN the system returns 201 with the created template (exercises can be added later)

#### Scenario: Create with invalid exercise ID

- GIVEN an exercise_id that does not exist in the catalog
- WHEN the user submits POST /templates
- THEN the system returns 400 with a validation error

### Requirement: List User Templates

The system MUST return the authenticated user's templates with cursor-based pagination.

#### Scenario: List templates

- GIVEN an authenticated user with 3 templates
- WHEN the user requests GET /templates
- THEN the system returns all 3 templates with their exercise slot counts

#### Scenario: User with no templates

- GIVEN an authenticated user with no templates
- WHEN the user requests GET /templates
- THEN the system returns 200 with an empty list

### Requirement: Get Template by ID

The system MUST return a single template with full exercise slot details.

#### Scenario: Own template

- GIVEN an authenticated user who owns template "t-123"
- WHEN the user requests GET /templates/t-123
- THEN the system returns 200 with template details and ordered exercise slots

#### Scenario: Another user's template

- GIVEN a template owned by a different user
- WHEN the user requests GET /templates/{other-user-template-id}
- THEN the system returns 404 (MUST NOT leak existence)

### Requirement: Update Template

The system MUST allow updating template name, description, and exercise slots. Updates replace the entire exercise slot list.

#### Scenario: Update template details

- GIVEN an authenticated user who owns template "t-123"
- WHEN the user submits PUT /templates/t-123 with updated name and exercises
- THEN the system returns 200 with the updated template

#### Scenario: Update another user's template

- GIVEN a template owned by a different user
- WHEN the user submits PUT /templates/{other-user-template-id}
- THEN the system returns 404

### Requirement: Delete Template

The system MUST soft-delete or hard-delete a template. Deleting a template MUST NOT affect existing sessions that were created from it.

#### Scenario: Delete own template

- GIVEN an authenticated user who owns template "t-123"
- WHEN the user submits DELETE /templates/t-123
- THEN the system removes the template and returns 204

#### Scenario: Delete another user's template

- GIVEN a template owned by a different user
- WHEN the user submits DELETE /templates/{other-user-template-id}
- THEN the system returns 404

## Constraints

- Templates are per-user (user_id scoping on all queries)
- Exercise slots MUST reference valid exercise IDs from the catalog
- Template deletion MUST NOT cascade-delete workout sessions
- Maximum 50 exercise slots per template (MVP limit)

## Dependencies

- auth (user authentication)
- exercise-catalog (exercise IDs must exist)
