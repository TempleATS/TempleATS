-- name: UpsertCandidate :one
INSERT INTO candidates (name, email, phone, resume_url, resume_filename, organization_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (email, organization_id) DO UPDATE
SET name = EXCLUDED.name, phone = EXCLUDED.phone,
    resume_url = COALESCE(EXCLUDED.resume_url, candidates.resume_url),
    resume_filename = COALESCE(EXCLUDED.resume_filename, candidates.resume_filename),
    updated_at = now()
RETURNING *;

-- name: GetCandidateByID :one
SELECT * FROM candidates WHERE id = $1 AND organization_id = $2;

-- name: ListCandidatesByOrg :many
SELECT * FROM candidates WHERE organization_id = $1 ORDER BY created_at DESC;

-- name: SearchCandidates :many
SELECT * FROM candidates
WHERE organization_id = $1
  AND (name ILIKE '%' || $2 || '%' OR email ILIKE '%' || $2 || '%')
ORDER BY created_at DESC;
