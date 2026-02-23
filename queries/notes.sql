-- name: CreateNote :one
INSERT INTO notes (content, application_id, author_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListNotesByApplication :many
SELECT n.*, u.name AS author_name
FROM notes n
JOIN users u ON n.author_id = u.id
WHERE n.application_id = $1
ORDER BY n.created_at;
