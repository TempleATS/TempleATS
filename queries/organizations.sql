-- name: CreateOrganization :one
INSERT INTO organizations (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, slug = $3, logo_url = $4, website = $5, updated_at = now()
WHERE id = $1
RETURNING *;
