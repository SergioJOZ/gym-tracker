-- Convert name/description to JSONB to support multi-language content.
-- Existing data is wrapped into {"en": <value>}.

DROP INDEX IF EXISTS idx_exercises_search;
ALTER TABLE exercises DROP COLUMN IF EXISTS search_vector;

-- Drop existing defaults before altering types
ALTER TABLE exercises ALTER COLUMN name DROP DEFAULT;
ALTER TABLE exercises ALTER COLUMN description DROP DEFAULT;

ALTER TABLE exercises
    ALTER COLUMN name TYPE JSONB USING jsonb_build_object('en', name),
    ALTER COLUMN description TYPE JSONB USING jsonb_build_object('en', description);

ALTER TABLE exercises
    ALTER COLUMN name SET NOT NULL,
    ALTER COLUMN name SET DEFAULT '{"en": ""}',
    ALTER COLUMN description SET NOT NULL,
    ALTER COLUMN description SET DEFAULT '{}';

ALTER TABLE exercises DROP CONSTRAINT IF EXISTS exercises_name_check;

ALTER TABLE exercises RENAME COLUMN name TO names;
ALTER TABLE exercises RENAME COLUMN description TO descriptions;

ALTER TABLE exercises ADD COLUMN search_vector tsvector NOT NULL
    GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(names->>'en', '')), 'A') ||
        setweight(to_tsvector('english', coalesce(descriptions->>'en', '')), 'B') ||
        setweight(to_tsvector('english', coalesce(muscle_group, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(category, '')), 'C')
    ) STORED;

CREATE INDEX idx_exercises_search ON exercises USING GIN (search_vector);
