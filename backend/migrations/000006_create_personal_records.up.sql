CREATE TABLE personal_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    max_weight  DECIMAL(8,2),
    max_reps    INT,
    max_volume  DECIMAL(10,2),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, exercise_id)
);

CREATE INDEX idx_personal_records_user_id ON personal_records(user_id);
