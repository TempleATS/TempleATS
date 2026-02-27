-- Structured job descriptions: replace single description with 4 sections
-- Also add org-level defaults for company blurb and closing statement

-- Add 4 new columns to jobs
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS company_blurb TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS team_details TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS skills TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS closing_statement TEXT NOT NULL DEFAULT '';

-- Migrate existing description content into team_details
UPDATE jobs SET team_details = description WHERE description IS NOT NULL AND description != '';

-- Drop the old description column
ALTER TABLE jobs DROP COLUMN IF EXISTS description;

-- Add org-level defaults
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS default_company_blurb TEXT NOT NULL DEFAULT '';
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS default_closing_statement TEXT NOT NULL DEFAULT '';
