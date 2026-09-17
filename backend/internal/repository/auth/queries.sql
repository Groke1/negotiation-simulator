-- name: CreateRefreshToken :one
INSERT INTO users.refresh_tokens (
    id,
    user_id,
    token_hash,
    active_role,
    expires_at
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(active_role),
    sqlc.arg(expires_at)
)
RETURNING id, user_id, token_hash, active_role, expires_at, revoked_at, created_at, updated_at;

-- name: GetRefreshTokenByHashForUpdate :one
SELECT id, user_id, token_hash, active_role, expires_at, revoked_at, created_at, updated_at
FROM users.refresh_tokens
WHERE token_hash = sqlc.arg(token_hash)
FOR UPDATE;

-- name: RotateRefreshToken :one
UPDATE users.refresh_tokens
SET token_hash = sqlc.arg(new_token_hash),
    expires_at = sqlc.arg(expires_at),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND token_hash = sqlc.arg(current_token_hash)
  AND revoked_at IS NULL
  AND expires_at > NOW()
RETURNING id, user_id, token_hash, active_role, expires_at, revoked_at, created_at, updated_at;

-- name: RevokeRefreshToken :exec
UPDATE users.refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: UpdateRefreshTokenActiveRole :one
UPDATE users.refresh_tokens
SET active_role = sqlc.arg(active_role), updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND revoked_at IS NULL
  AND expires_at > NOW()
RETURNING id, user_id, token_hash, active_role, expires_at, revoked_at, created_at, updated_at;

-- name: DeleteExpiredTokens :exec
DELETE FROM users.refresh_tokens WHERE expires_at < NOW();
