-- TempleATS Schema
-- Multi-tenant ATS with requisitions, pipeline stages, and audit trail

-- Organizations (tenants)
CREATE TABLE organizations (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    logo_url    TEXT,
    website     TEXT,
    default_company_blurb      TEXT NOT NULL DEFAULT '',
    default_closing_statement  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Users (org members)
CREATE TABLE users (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email           TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'interviewer',
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_org ON users(organization_id);

-- Requisitions (hiring needs)
CREATE TABLE requisitions (
    id                TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    title             TEXT NOT NULL,
    job_code          TEXT,              -- e.g., SWE, MLE, SYS
    level             TEXT,
    department        TEXT,
    target_hires      INT NOT NULL DEFAULT 1,
    status            TEXT NOT NULL DEFAULT 'open',
    hiring_manager_id TEXT REFERENCES users(id),
    organization_id   TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    opened_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at         TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_reqs_org ON requisitions(organization_id);
CREATE INDEX idx_reqs_status ON requisitions(status);

-- Jobs (postings attached to optional requisition)
CREATE TABLE jobs (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    title           TEXT NOT NULL,
    company_blurb     TEXT NOT NULL DEFAULT '',
    team_details      TEXT NOT NULL DEFAULT '',
    responsibilities  TEXT NOT NULL DEFAULT '',
    qualifications    TEXT NOT NULL DEFAULT '',
    closing_statement TEXT NOT NULL DEFAULT '',
    location        TEXT,
    department      TEXT,
    salary          TEXT,
    status          TEXT NOT NULL DEFAULT 'draft',
    requisition_id  TEXT REFERENCES requisitions(id) ON DELETE SET NULL,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_jobs_org ON jobs(organization_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_req ON jobs(requisition_id);

-- Candidates (applicants, deduped by email+org)
CREATE TABLE candidates (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name            TEXT NOT NULL,
    email           TEXT NOT NULL,
    phone           TEXT,
    resume_url      TEXT,
    resume_filename TEXT,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(email, organization_id)
);
CREATE INDEX idx_candidates_org ON candidates(organization_id);

-- Applications (candidate <-> job, with pipeline stage)
CREATE TABLE applications (
    id               TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    stage            TEXT NOT NULL DEFAULT 'applied',
    rejection_reason TEXT,
    rejection_notes  TEXT,
    candidate_id     TEXT NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    job_id           TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(candidate_id, job_id)
);
CREATE INDEX idx_apps_job ON applications(job_id);
CREATE INDEX idx_apps_stage ON applications(stage);

-- Stage transitions (audit trail)
CREATE TABLE stage_transitions (
    id             TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    from_stage     TEXT,
    to_stage       TEXT NOT NULL,
    moved_by_id    TEXT REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_transitions_app ON stage_transitions(application_id);

-- Notes (on applications)
CREATE TABLE notes (
    id             TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    content        TEXT NOT NULL,
    application_id TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    author_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notes_app ON notes(application_id);

-- Interview assignments (per-application)
CREATE TABLE interview_assignments (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    application_id  TEXT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    interviewer_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(application_id, interviewer_id)
);
CREATE INDEX idx_interview_assignments_app ON interview_assignments(application_id);
CREATE INDEX idx_interview_assignments_user ON interview_assignments(interviewer_id);

-- Invitations (for adding team members)
CREATE TABLE invitations (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email           TEXT NOT NULL,
    role            TEXT NOT NULL DEFAULT 'interviewer',
    token           TEXT NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    expires_at      TIMESTAMPTZ NOT NULL,
    accepted_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_invitations_org ON invitations(organization_id);
CREATE INDEX idx_invitations_token ON invitations(token);
