-- name: CreateInvitation :one
INSERT INTO invitations (email, role, organization_id, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetInvitationByToken :one
SELECT * FROM invitations WHERE token = $1 AND accepted_at IS NULL AND expires_at > now();

-- name: AcceptInvitation :one
UPDATE invitations SET accepted_at = now() WHERE id = $1 RETURNING *;

-- name: ListInvitationsByOrg :many
SELECT * FROM invitations WHERE organization_id = $1 ORDER BY created_at DESC;
