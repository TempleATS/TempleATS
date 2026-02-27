-- Add recruiter_id to requisitions for tracking which recruiter owns the req
ALTER TABLE requisitions ADD COLUMN IF NOT EXISTS recruiter_id TEXT REFERENCES users(id) ON DELETE SET NULL;
