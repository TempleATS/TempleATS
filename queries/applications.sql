-- name: CreateApplication :one
INSERT INTO applications (stage, candidate_id, job_id)
VALUES ('applied', $1, $2)
RETURNING *;

-- name: GetApplicationByID :one
SELECT * FROM applications WHERE id = $1;

-- name: GetApplicationWithDetails :one
SELECT a.*,
       c.name AS candidate_name, c.email AS candidate_email, c.phone AS candidate_phone,
       c.resume_url AS candidate_resume_url,
       j.title AS job_title
FROM applications a
JOIN candidates c ON a.candidate_id = c.id
JOIN jobs j ON a.job_id = j.id
WHERE a.id = $1;

-- name: ListApplicationsByJob :many
SELECT a.*,
       c.name AS candidate_name, c.email AS candidate_email,
       c.resume_url AS candidate_resume_url
FROM applications a
JOIN candidates c ON a.candidate_id = c.id
WHERE a.job_id = $1
ORDER BY a.created_at DESC;

-- name: ListApplicationsByCandidate :many
SELECT a.*, j.title AS job_title, j.status AS job_status
FROM applications a
JOIN jobs j ON a.job_id = j.id
WHERE a.candidate_id = $1
ORDER BY a.created_at DESC;

-- name: UpdateApplicationStage :one
UPDATE applications
SET stage = $2, rejection_reason = $3, rejection_notes = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CountApplicationsByJobAndStage :many
SELECT stage, COUNT(*)::int AS count
FROM applications
WHERE job_id = $1
GROUP BY stage;

-- name: FunnelByRequisition :many
SELECT a.stage, COUNT(*)::int AS count
FROM applications a
JOIN jobs j ON a.job_id = j.id
WHERE j.requisition_id = $1
GROUP BY a.stage;

-- name: RejectionsByRequisition :many
SELECT a.rejection_reason, COUNT(*)::int AS count
FROM applications a
JOIN jobs j ON a.job_id = j.id
WHERE j.requisition_id = $1 AND a.stage = 'rejected' AND a.rejection_reason IS NOT NULL
GROUP BY a.rejection_reason;

-- name: ByJobBreakdown :many
SELECT j.id AS job_id, j.title AS job_title,
       COUNT(*)::int AS total,
       COUNT(*) FILTER (WHERE a.stage = 'hired')::int AS hired,
       COUNT(*) FILTER (WHERE a.stage = 'rejected')::int AS rejected
FROM applications a
JOIN jobs j ON a.job_id = j.id
WHERE j.requisition_id = $1
GROUP BY j.id, j.title;
