-- Revert: convert JSONB names/descriptions back to plain text (English only).

DROP INDEX IF EXISTS idx_exercises_search;
ALTER TABLE exercises DROP COLUMN IF EXISTS search_vector;

ALTER TABLE exercises RENAME COLUMN names TO name;
ALTER TABLE exercises RENAME COLUMN descriptions TO description;

-- Drop JSONB defaults before converting types
ALTER TABLE exercises ALTER COLUMN name DROP DEFAULT;
ALTER TABLE exercises ALTER COLUMN description DROP DEFAULT;

ALTER TABLE exercises
    ALTER COLUMN name TYPE VARCHAR(255) USING name->>'en',
    ALTER COLUMN description TYPE TEXT USING COALESCE(description->>'en', '');

ALTER TABLE exercises
    ALTER COLUMN name SET NOT NULL,
    ALTER COLUMN description SET NOT NULL,
    ALTER COLUMN description SET DEFAULT '';

ALTER TABLE exercises ADD COLUMN search_vector tsvector NOT NULL
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(muscle_group, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(category, '')), 'C')
    ) STORED;

CREATE INDEX idx_exercises_search ON exercises USING GIN (search_vector);
