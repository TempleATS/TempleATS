-- name: CreateJob :one
INSERT INTO jobs (title, company_blurb, team_details, responsibilities, qualifications, closing_statement, location, department, salary, status, requisition_id, organization_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetJobByID :one
SELECT * FROM jobs WHERE id = $1 AND organization_id = $2;

-- name: GetJobPublic :one
SELECT * FROM jobs WHERE id = $1 AND status = 'open';

-- name: ListJobsByOrg :many
SELECT * FROM jobs WHERE organization_id = $1 ORDER BY created_at DESC;

-- name: ListOpenJobsByOrgSlug :many
SELECT j.* FROM jobs j
JOIN organizations o ON j.organization_id = o.id
WHERE o.slug = $1 AND j.status = 'open'
ORDER BY j.created_at DESC;

-- name: ListJobsByRequisition :many
SELECT * FROM jobs WHERE requisition_id = $1 AND organization_id = $2 ORDER BY created_at DESC;

-- name: ListJobsByHiringManager :many
SELECT j.* FROM jobs j
JOIN requisitions r ON j.requisition_id = r.id
WHERE j.organization_id = $1 AND r.hiring_manager_id = $2
ORDER BY j.created_at DESC;

-- name: ListJobsByInterviewer :many
SELECT DISTINCT j.* FROM jobs j
JOIN applications a ON a.job_id = j.id
JOIN interview_assignments ia ON ia.application_id = a.id
WHERE j.organization_id = $1 AND ia.interviewer_id = $2
ORDER BY j.created_at DESC;

-- name: UpdateJob :one
UPDATE jobs
SET title = $2, company_blurb = $3, team_details = $4, responsibilities = $5, qualifications = $6,
    closing_statement = $7, location = $8, department = $9, salary = $10, status = $11,
    requisition_id = $12, updated_at = now()
WHERE id = $1 AND organization_id = $13
RETURNING *;
