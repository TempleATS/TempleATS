CREATE TABLE IF NOT EXISTS interview_feedback (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    application_id  TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    stage           TEXT NOT NULL,
    recommendation  TEXT NOT NULL DEFAULT 'none',
    content         TEXT NOT NULL,
    author_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
