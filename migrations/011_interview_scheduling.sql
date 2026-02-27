-- OAuth calendar connections per user
CREATE TABLE user_calendar_connections (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL CHECK (provider IN ('google', 'microsoft')),
    access_token    TEXT NOT NULL,
    refresh_token   TEXT NOT NULL,
    token_expiry    TIMESTAMPTZ NOT NULL,
    calendar_email  TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, provider)
);

-- Interview schedule (one per scheduling action)
CREATE TABLE interview_schedules (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    application_id  TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    token           TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    duration_minutes INT NOT NULL DEFAULT 60,
    location        TEXT,
    meeting_url     TEXT,
    notes           TEXT,
    created_by      TEXT NOT NULL REFERENCES users(id),
    confirmed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_interview_schedules_app ON interview_schedules(application_id);
CREATE INDEX idx_interview_schedules_token ON interview_schedules(token);

-- Proposed time slots for a schedule
CREATE TABLE interview_slots (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    schedule_id     TEXT NOT NULL REFERENCES interview_schedules(id) ON DELETE CASCADE,
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    selected        BOOLEAN NOT NULL DEFAULT false
);
CREATE INDEX idx_interview_slots_schedule ON interview_slots(schedule_id);

-- Which interviewers are part of a schedule
CREATE TABLE interview_schedule_interviewers (
    schedule_id     TEXT NOT NULL REFERENCES interview_schedules(id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    calendar_event_id TEXT,
    PRIMARY KEY (schedule_id, user_id)
);
