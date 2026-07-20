CREATE TABLE exercises (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    muscle_group    VARCHAR(100) NOT NULL,
    equipment       VARCHAR(100) NOT NULL DEFAULT '',
    difficulty      VARCHAR(50) NOT NULL DEFAULT 'beginner',
    category        VARCHAR(100) NOT NULL DEFAULT '',
    gif_path        VARCHAR(500) NOT NULL DEFAULT '',
    thumbnail_path  VARCHAR(500) NOT NULL DEFAULT '',
    search_vector   tsvector NOT NULL GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(muscle_group, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(category, '')), 'C')
    ) STORED,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exercises_search ON exercises USING GIN (search_vector);
CREATE INDEX idx_exercises_muscle_group ON exercises(muscle_group);
CREATE INDEX idx_exercises_equipment ON exercises(equipment);
CREATE INDEX idx_exercises_difficulty ON exercises(difficulty);
CREATE INDEX idx_exercises_category ON exercises(category);
CREATE INDEX idx_exercises_name ON exercises(name);
