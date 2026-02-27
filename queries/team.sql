-- name: UpdateUserRole :one
UPDATE users SET role = $2, updated_at = now()
WHERE id = $1 AND organization_id = $3
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1 AND organization_id = $2;
