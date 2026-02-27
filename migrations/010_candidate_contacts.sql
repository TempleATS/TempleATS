CREATE TABLE candidate_contacts (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    candidate_id    TEXT NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    category        TEXT NOT NULL CHECK (category IN ('email', 'phone', 'online_presence')),
    label           TEXT NOT NULL,
    value           TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(candidate_id, category, value)
);
CREATE INDEX idx_candidate_contacts_candidate ON candidate_contacts(candidate_id);
