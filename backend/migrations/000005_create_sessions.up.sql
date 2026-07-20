CREATE TABLE workout_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    template_id UUID REFERENCES workout_templates(id) ON DELETE SET NULL,
    name        VARCHAR(255) NOT NULL,
    notes       TEXT NOT NULL DEFAULT '',
    start_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    end_at      TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_workout_sessions_user_id ON workout_sessions(user_id);
CREATE INDEX idx_workout_sessions_start_at ON workout_sessions(start_at DESC);

CREATE TABLE session_exercises (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    "order"     INT NOT NULL DEFAULT 0,
    notes       TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_session_exercises_session_id ON session_exercises(session_id);
CREATE INDEX idx_session_exercises_exercise_id ON session_exercises(exercise_id);

CREATE TABLE session_sets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_exercise_id UUID NOT NULL REFERENCES session_exercises(id) ON DELETE CASCADE,
    "order"             INT NOT NULL DEFAULT 0,
    weight              DECIMAL(8,2),
    reps                INT,
    duration            INT, -- duration in seconds
    rpe                 DECIMAL(3,1)
);

CREATE INDEX idx_session_sets_session_exercise_id ON session_sets(session_exercise_id);
