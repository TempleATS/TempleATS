-- SMTP settings per organization
CREATE TABLE IF NOT EXISTS smtp_settings (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organization_id TEXT NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
    host            TEXT NOT NULL,
    port            INT NOT NULL DEFAULT 587,
    username        TEXT NOT NULL,
    password        TEXT NOT NULL,
    from_email      TEXT NOT NULL,
    from_name       TEXT NOT NULL DEFAULT '',
    tls_enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Configurable email templates for stage transitions
CREATE TABLE IF NOT EXISTS email_templates (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    stage           TEXT NOT NULL,
    subject         TEXT NOT NULL,
    body            TEXT NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(organization_id, stage)
);

-- Audit log for all sent emails
CREATE TABLE IF NOT EXISTS notifications (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type            TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    recipient_name  TEXT NOT NULL DEFAULT '',
    subject         TEXT NOT NULL,
    body            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'sent',
    error_message   TEXT,
    application_id  TEXT REFERENCES applications(id) ON DELETE SET NULL,
    note_id         TEXT REFERENCES notes(id) ON DELETE SET NULL,
    triggered_by_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_notifications_org ON notifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_notifications_app ON notifications(application_id);
