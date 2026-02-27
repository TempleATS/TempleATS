-- Referrals table: tracks both direct referrals and referral links
CREATE TABLE referrals (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    referrer_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    job_id          TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    token           TEXT NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    source          TEXT NOT NULL CHECK (source IN ('direct', 'link')),
    candidate_name  TEXT,
    candidate_id    TEXT REFERENCES candidates(id) ON DELETE SET NULL,
    application_id  TEXT REFERENCES applications(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_referrals_org ON referrals(organization_id);
CREATE INDEX idx_referrals_referrer ON referrals(referrer_id);
CREATE INDEX idx_referrals_token ON referrals(token);

-- Link applications back to referrals
ALTER TABLE applications ADD COLUMN referral_id TEXT REFERENCES referrals(id) ON DELETE SET NULL;
