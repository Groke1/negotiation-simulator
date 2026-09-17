-- name: CreateOrganization :one
INSERT INTO organizations.organizations (id, name)
VALUES (sqlc.arg(id), sqlc.arg(name))
RETURNING id, name, created_at;

-- name: GetOrganizationByID :one
SELECT id, name, created_at
FROM organizations.organizations
WHERE id = sqlc.arg(id);

-- name: ListOrganizations :many
SELECT id, name, created_at
FROM organizations.organizations
WHERE sqlc.arg(query)::text = '' OR name ILIKE '%' || sqlc.arg(query)::text || '%'
ORDER BY name ASC, id ASC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);
