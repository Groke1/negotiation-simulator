-- name: CreateUser :one
INSERT INTO users.users (id, username, email, password_hash)
VALUES (sqlc.arg(id), sqlc.arg(username), LOWER(sqlc.arg(email)), sqlc.arg(password_hash))
RETURNING id, organization_id, username, email, password_hash, created_at, updated_at;

-- name: ListProfile :many
WITH seeded AS (
    INSERT INTO learning.user_profile (user_id, cluster_id, score, evidence_mass, confidence)
    SELECT sqlc.arg(user_id), id, 50, 0, 0 FROM learning.competency_clusters
    ON CONFLICT (user_id, cluster_id) DO NOTHING
)
SELECT c.id AS cluster_id, c.code, c.name,
       FLOOR(p.score)::integer AS score,
       p.score AS score_raw,
       FLOOR(p.confidence)::integer AS confidence,
       p.confidence AS confidence_raw,
       p.evidence_mass,
       p.updated_at
FROM learning.competency_clusters c
JOIN learning.user_profile p ON p.cluster_id = c.id
WHERE p.user_id = sqlc.arg(user_id)
ORDER BY c.code;

-- name: ListProfileHistory :many
SELECT id, source, source_id, changes, created_at
FROM learning.profile_history
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.narg(from_at)::timestamptz IS NULL OR created_at >= sqlc.narg(from_at))
  AND (sqlc.narg(to_at)::timestamptz IS NULL OR created_at <= sqlc.narg(to_at))
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: ListOrganizationUsers :many
SELECT u.id, u.username, u.email, COALESCE(array_agg(ur.role ORDER BY ur.role) FILTER (WHERE ur.role IS NOT NULL), '{}')::users.product_role[] AS roles, u.created_at
FROM users.users requester
JOIN users.users u ON u.organization_id = requester.organization_id
LEFT JOIN users.user_roles ur ON ur.user_id = u.id
WHERE requester.id = sqlc.arg(requester_id)
GROUP BY u.id, u.username, u.email, u.created_at
ORDER BY u.created_at DESC, u.id DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: GetUserByEmail :one
SELECT id, organization_id, username, email, password_hash, created_at, updated_at
FROM users.users
WHERE LOWER(email) = LOWER(sqlc.arg(email));

-- name: GetUserByID :one
SELECT id, organization_id, username, email, password_hash, created_at, updated_at
FROM users.users
WHERE id = sqlc.arg(id);

-- name: AddUserRole :exec
INSERT INTO users.user_roles (user_id, role)
VALUES (sqlc.arg(user_id), sqlc.arg(role));

-- name: GetUserRoles :many
SELECT role
FROM users.user_roles
WHERE user_id = sqlc.arg(user_id)
ORDER BY role;

-- name: SetUserOrganization :one
UPDATE users.users
SET organization_id = sqlc.arg(organization_id), updated_at = NOW()
WHERE id = sqlc.arg(user_id)
  AND organization_id IS NULL
RETURNING id, organization_id, username, email, password_hash, created_at, updated_at;
