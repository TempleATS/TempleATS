-- RBAC migration: interview assignments, role updates, pipeline stage updates

-- Interview assignments (per-application)
CREATE TABLE IF NOT EXISTS interview_assignments (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    application_id  TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    interviewer_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(application_id, interviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_interview_assignments_app ON interview_assignments(application_id);
CREATE INDEX IF NOT EXISTS idx_interview_assignments_user ON interview_assignments(interviewer_id);

-- Migrate existing roles
UPDATE users SET role = 'super_admin' WHERE role = 'admin';
UPDATE users SET role = 'interviewer' WHERE role = 'recruiter';

-- Migrate pipeline stages
UPDATE applications SET stage = 'hm_review' WHERE stage = 'screening';
UPDATE applications SET stage = 'first_interview' WHERE stage = 'interview';

-- Update stage_transitions history
UPDATE stage_transitions SET from_stage = 'hm_review' WHERE from_stage = 'screening';
UPDATE stage_transitions SET to_stage = 'hm_review' WHERE to_stage = 'screening';
UPDATE stage_transitions SET from_stage = 'first_interview' WHERE from_stage = 'interview';
UPDATE stage_transitions SET to_stage = 'first_interview' WHERE to_stage = 'interview';

-- Update default role on users and invitations tables
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'interviewer';
ALTER TABLE invitations ALTER COLUMN role SET DEFAULT 'interviewer';
