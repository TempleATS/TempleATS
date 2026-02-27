-- name: CreateRequisition :one
INSERT INTO requisitions (title, job_code, level, department, target_hires, hiring_manager_id, recruiter_id, organization_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetRequisitionByID :one
SELECT * FROM requisitions WHERE id = $1 AND organization_id = $2;

-- name: ListRequisitionsByOrg :many
SELECT * FROM requisitions WHERE organization_id = $1 ORDER BY created_at DESC;

-- name: ListRequisitionsByHiringManager :many
SELECT * FROM requisitions
WHERE organization_id = $1 AND hiring_manager_id = $2
ORDER BY created_at DESC;

-- name: UpdateRequisition :one
UPDATE requisitions
SET title = $2, job_code = $3, level = $4, department = $5, target_hires = $6,
    status = $7, hiring_manager_id = $8, recruiter_id = $9, closed_at = $10, updated_at = now()
WHERE id = $1 AND organization_id = $11
RETURNING *;

-- name: DeleteRequisition :exec
DELETE FROM requisitions WHERE id = $1 AND organization_id = $2;

-- name: CountHiredByRequisition :one
SELECT COUNT(*) FROM applications a
JOIN jobs j ON a.job_id = j.id
WHERE j.requisition_id = $1 AND a.stage = 'hired';
