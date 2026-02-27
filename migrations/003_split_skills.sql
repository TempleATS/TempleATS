-- Split skills column into responsibilities and qualifications
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS responsibilities TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS qualifications TEXT NOT NULL DEFAULT '';
UPDATE jobs SET responsibilities = skills WHERE skills != '';
ALTER TABLE jobs DROP COLUMN IF EXISTS skills;
