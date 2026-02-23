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

-- name: AvgTimeInStageByRequisition :many
SELECT st.from_stage AS stage,
       AVG(EXTRACT(EPOCH FROM (st.created_at - prev.created_at)) / 86400)::float AS avg_days
FROM stage_transitions st
JOIN applications a ON st.application_id = a.id
JOIN jobs j ON a.job_id = j.id
JOIN stage_transitions prev ON prev.application_id = st.application_id
  AND prev.to_stage = st.from_stage
WHERE j.requisition_id = $1 AND st.from_stage IS NOT NULL
GROUP BY st.from_stage;
