-- name: CreateStageTransition :one
INSERT INTO stage_transitions (application_id, from_stage, to_stage, moved_by_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListTransitionsByApplication :many
SELECT st.*, u.name AS moved_by_name
FROM stage_transitions st
LEFT JOIN users u ON st.moved_by_id = u.id
WHERE st.application_id = $1
ORDER BY st.created_at;
