-- name: CreateMedia :one
INSERT INTO learning.media (id, owner_id, media_type, content_type, payload, duration_seconds)
VALUES (sqlc.arg(id), sqlc.arg(owner_id), sqlc.arg(media_type), sqlc.arg(content_type), sqlc.arg(payload), sqlc.narg(duration_seconds))
RETURNING id, owner_id, media_type, content_type, duration_seconds, created_at;

-- name: GetMedia :one
SELECT id, owner_id, media_type, content_type, duration_seconds, created_at
FROM learning.media
WHERE id = sqlc.arg(id);

-- name: GetMediaContent :one
SELECT owner_id, content_type, payload
FROM learning.media
WHERE id = sqlc.arg(id);
