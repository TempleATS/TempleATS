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

-- name: ListCandidatesByHiringManager :many
SELECT DISTINCT c.* FROM candidates c
JOIN applications a ON a.candidate_id = c.id
JOIN jobs j ON a.job_id = j.id
JOIN requisitions r ON j.requisition_id = r.id
WHERE c.organization_id = $1 AND r.hiring_manager_id = $2
ORDER BY c.created_at DESC;

-- name: ListCandidatesByInterviewer :many
SELECT DISTINCT c.* FROM candidates c
JOIN applications a ON a.candidate_id = c.id
JOIN interview_assignments ia ON ia.application_id = a.id
WHERE c.organization_id = $1 AND ia.interviewer_id = $2
ORDER BY c.created_at DESC;

-- name: SearchCandidates :many
SELECT * FROM candidates
WHERE organization_id = $1
  AND (name ILIKE '%' || $2 || '%' OR email ILIKE '%' || $2 || '%')
ORDER BY created_at DESC;

-- name: ListCandidatesWithAllApps :many
SELECT c.id, c.name, c.email, c.phone, c.resume_url, c.resume_filename, c.created_at,
       COALESCE(a.id, '') AS app_id, COALESCE(a.stage, '') AS app_stage,
       COALESCE(j.title, '') AS job_title, a.created_at AS applied_at
FROM candidates c
LEFT JOIN applications a ON a.candidate_id = c.id
LEFT JOIN jobs j ON a.job_id = j.id
WHERE c.organization_id = $1
ORDER BY c.name ASC, a.created_at DESC;

-- name: ListCandidatesWithLatestAppByHM :many
SELECT c.id, c.name, c.email, c.phone, c.resume_url, c.resume_filename, c.created_at,
       a.id AS app_id, a.stage AS app_stage, j.title AS job_title, a.created_at AS applied_at
FROM candidates c
JOIN applications a ON a.candidate_id = c.id
JOIN jobs j ON a.job_id = j.id
JOIN requisitions r ON j.requisition_id = r.id
WHERE c.organization_id = $1 AND r.hiring_manager_id = $2
ORDER BY a.created_at DESC;

-- name: ListCandidatesWithLatestAppByInterviewer :many
SELECT c.id, c.name, c.email, c.phone, c.resume_url, c.resume_filename, c.created_at,
       a.id AS app_id, a.stage AS app_stage, j.title AS job_title, a.created_at AS applied_at
FROM candidates c
JOIN applications a ON a.candidate_id = c.id
JOIN interview_assignments ia ON ia.application_id = a.id
JOIN jobs j ON a.job_id = j.id
WHERE c.organization_id = $1 AND ia.interviewer_id = $2
ORDER BY a.created_at DESC;

-- name: UpdateCandidateContact :one
UPDATE candidates
SET email = $3, phone = $4, linkedin_url = $5, updated_at = now()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- name: SearchCandidatesWithAllApps :many
SELECT c.id, c.name, c.email, c.phone, c.resume_url, c.resume_filename, c.created_at,
       COALESCE(a.id, '') AS app_id, COALESCE(a.stage, '') AS app_stage,
       COALESCE(j.title, '') AS job_title, a.created_at AS applied_at
FROM candidates c
LEFT JOIN applications a ON a.candidate_id = c.id
LEFT JOIN jobs j ON a.job_id = j.id
WHERE c.organization_id = $1
  AND (c.name ILIKE '%' || $2 || '%' OR c.email ILIKE '%' || $2 || '%')
ORDER BY c.name ASC, a.created_at DESC;
