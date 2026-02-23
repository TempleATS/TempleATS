-- name: CreateJob :one
INSERT INTO jobs (title, description, location, department, salary, status, requisition_id, organization_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

-- name: UpdateJob :one
UPDATE jobs
SET title = $2, description = $3, location = $4, department = $5,
    salary = $6, status = $7, requisition_id = $8, updated_at = now()
WHERE id = $1 AND organization_id = $9
RETURNING *;
