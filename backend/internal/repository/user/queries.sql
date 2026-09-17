-- name: CreateUser :one
INSERT INTO users.users (id, username, email, password_hash)
VALUES (sqlc.arg(id), sqlc.arg(username), LOWER(sqlc.arg(email)), sqlc.arg(password_hash))
RETURNING id, organization_id, username, email, password_hash, created_at, updated_at;

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
