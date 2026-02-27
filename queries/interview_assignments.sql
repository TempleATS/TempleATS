-- name: CreateInterviewAssignment :one
INSERT INTO interview_assignments (application_id, interviewer_id)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteInterviewAssignment :exec
DELETE FROM interview_assignments
WHERE application_id = $1 AND interviewer_id = $2;

-- name: ListInterviewersByApplication :many
SELECT ia.*, u.name AS interviewer_name, u.email AS interviewer_email
FROM interview_assignments ia
JOIN users u ON ia.interviewer_id = u.id
WHERE ia.application_id = $1
ORDER BY ia.created_at;

-- name: ListApplicationsByInterviewer :many
SELECT a.*, c.name AS candidate_name, c.email AS candidate_email,
       c.resume_url AS candidate_resume_url,
       j.title AS job_title, j.id AS job_id
FROM interview_assignments ia
JOIN applications a ON ia.application_id = a.id
JOIN candidates c ON a.candidate_id = c.id
JOIN jobs j ON a.job_id = j.id
WHERE ia.interviewer_id = $1 AND j.organization_id = $2
ORDER BY a.created_at DESC;
