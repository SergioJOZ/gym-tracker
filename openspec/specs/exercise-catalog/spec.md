# Exercise Catalog Specification

## Purpose

Read-only browsing, searching, and filtering of a seeded exercise dataset (1,324 exercises). No user-created exercises in MVP.

## Requirements

### Requirement: List Exercises

The system MUST return exercises with cursor-based pagination. Default sort by name ascending.

#### Scenario: List first page

- GIVEN the exercise catalog is seeded
- WHEN the user requests GET /exercises with limit=20
- THEN the system returns up to 20 exercises with a next_cursor
- AND each exercise includes id, name, category, muscle_group, equipment

#### Scenario: List next page

- GIVEN a next_cursor from a previous response
- WHEN the user requests GET /exercises?cursor={value}&limit=20
- THEN the system returns the next 20 exercises after that cursor

#### Scenario: Empty catalog

- GIVEN the catalog has not been seeded
- WHEN the user requests GET /exercises
- THEN the system returns 200 with an empty list

### Requirement: Search Exercises by Name

The system MUST support full-text search on exercise name.

#### Scenario: Search with matching results

- GIVEN exercises exist with "bench press" in the name
- WHEN the user requests GET /exercises?search=bench
- THEN the system returns matching exercises ordered by relevance

#### Scenario: Search with no results

- GIVEN no exercises match the search term
- WHEN the user requests GET /exercises?search=xyznonexistent
- THEN the system returns 200 with an empty list

### Requirement: Filter Exercises

The system MUST support filtering by category, body_part, equipment, and target_muscle. Multiple filters MUST combine with AND logic.

#### Scenario: Single filter

- GIVEN exercises with category "strength"
- WHEN the user requests GET /exercises?category=strength
- THEN the system returns only strength exercises

#### Scenario: Multiple filters combined

- GIVEN exercises matching category=strength AND equipment=barbell
- WHEN the user requests GET /exercises?category=strength&equipment=barbell
- THEN the system returns only exercises matching both filters

#### Scenario: Invalid filter value

- GIVEN a filter value that matches no exercises
- WHEN the user requests GET /exercises?category=nonexistent
- THEN the system returns 200 with an empty list

### Requirement: Get Exercise by ID

The system MUST return a single exercise with full details.

#### Scenario: Existing exercise

- GIVEN an exercise with id "abc-123"
- WHEN the user requests GET /exercises/abc-123
- THEN the system returns 200 with full exercise details including gif_url and thumbnail_url

#### Scenario: Non-existing exercise

- GIVEN no exercise with id "xyz-999"
- WHEN the user requests GET /exercises/xyz-999
- THEN the system returns 404

## Constraints

- Catalog is READ-ONLY — no POST/PUT/DELETE for exercises
- Cursor-based pagination MUST use keyset pattern (not offset)
- Search MUST use PostgreSQL full-text search (GIN index on tsvector)
- Response time for search/filter MUST be < 100ms on 1,324 rows
- Exercise data is seeded once via CLI command from exercises-dataset

## Dependencies

- media-serving (gif_url and thumbnail_url reference static files)
