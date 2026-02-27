-- name: ListContactsByCandidate :many
SELECT * FROM candidate_contacts WHERE candidate_id = $1 ORDER BY category, label;

-- name: CreateCandidateContact :one
INSERT INTO candidate_contacts (candidate_id, category, label, value)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteCandidateContact :exec
DELETE FROM candidate_contacts WHERE id = $1 AND candidate_id = $2;
