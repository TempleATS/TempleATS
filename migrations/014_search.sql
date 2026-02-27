-- Enable pg_trgm for fuzzy matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add tsvector column to jobs for full-text search
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Populate existing rows
UPDATE jobs SET search_vector =
  setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
  setweight(to_tsvector('english', coalesce(location, '')), 'B') ||
  setweight(to_tsvector('english', coalesce(department, '')), 'B') ||
  setweight(to_tsvector('english', coalesce(responsibilities, '')), 'C') ||
  setweight(to_tsvector('english', coalesce(qualifications, '')), 'C');

-- GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_jobs_search ON jobs USING gin(search_vector);

-- Trigram index on title for fuzzy matching
CREATE INDEX IF NOT EXISTS idx_jobs_title_trgm ON jobs USING gin(title gin_trgm_ops);

-- Trigger to keep search_vector in sync
CREATE OR REPLACE FUNCTION jobs_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.location, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.department, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.responsibilities, '')), 'C') ||
    setweight(to_tsvector('english', coalesce(NEW.qualifications, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_jobs_search_vector ON jobs;
CREATE TRIGGER trg_jobs_search_vector
  BEFORE INSERT OR UPDATE OF title, location, department, responsibilities, qualifications
  ON jobs FOR EACH ROW
  EXECUTE FUNCTION jobs_search_vector_update();

-- Add tsvector column to requisitions
ALTER TABLE requisitions ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Populate existing rows
UPDATE requisitions SET search_vector =
  setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
  setweight(to_tsvector('english', coalesce(job_code, '')), 'B') ||
  setweight(to_tsvector('english', coalesce(department, '')), 'B') ||
  setweight(to_tsvector('english', coalesce(level, '')), 'B');

-- GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_reqs_search ON requisitions USING gin(search_vector);

-- Trigram index on title for fuzzy matching
CREATE INDEX IF NOT EXISTS idx_reqs_title_trgm ON requisitions USING gin(title gin_trgm_ops);

-- Trigger to keep search_vector in sync
CREATE OR REPLACE FUNCTION reqs_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.job_code, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.department, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.level, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_reqs_search_vector ON requisitions;
CREATE TRIGGER trg_reqs_search_vector
  BEFORE INSERT OR UPDATE OF title, job_code, department, level
  ON requisitions FOR EACH ROW
  EXECUTE FUNCTION reqs_search_vector_update();
