-- name: CreateUser :one
INSERT INTO users (email, name, password_hash, role, organization_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ListUsersByOrg :many
SELECT * FROM users WHERE organization_id = $1 ORDER BY created_at;
