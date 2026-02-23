-- name: CreateRequisition :one
INSERT INTO requisitions (title, level, department, target_hires, hiring_manager_id, organization_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRequisitionByID :one
SELECT * FROM requisitions WHERE id = $1 AND organization_id = $2;

-- name: ListRequisitionsByOrg :many
SELECT * FROM requisitions WHERE organization_id = $1 ORDER BY created_at DESC;

-- name: UpdateRequisition :one
UPDATE requisitions
SET title = $2, level = $3, department = $4, target_hires = $5,
    status = $6, hiring_manager_id = $7, closed_at = $8, updated_at = now()
WHERE id = $1 AND organization_id = $9
RETURNING *;

-- name: CountHiredByRequisition :one
SELECT COUNT(*) FROM applications a
JOIN jobs j ON a.job_id = j.id
WHERE j.requisition_id = $1 AND a.stage = 'hired';
