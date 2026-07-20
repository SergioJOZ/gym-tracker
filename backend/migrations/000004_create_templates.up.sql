CREATE TABLE workout_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_workout_templates_user_id ON workout_templates(user_id);

CREATE TABLE template_slots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id     UUID NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    exercise_id     UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    "order"         INT NOT NULL DEFAULT 0,
    target_sets     INT NOT NULL DEFAULT 0,
    target_reps     INT NOT NULL DEFAULT 0,
    target_weight   DECIMAL(8,2),
    target_duration INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_template_slots_template_id ON template_slots(template_id);
