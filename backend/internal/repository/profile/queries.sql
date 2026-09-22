-- name: ListCompetencyClusters :many
SELECT id, code, name, description
FROM learning.competency_clusters
ORDER BY code;

-- name: ListProfileForUpdate :many
WITH seeded AS (
    INSERT INTO learning.user_profile (user_id, cluster_id, score)
    SELECT sqlc.arg(user_id), id, 50 FROM learning.competency_clusters
    ON CONFLICT (user_id, cluster_id) DO NOTHING
)
SELECT c.id AS cluster_id, c.code, c.name, p.score::integer AS score, p.updated_at
FROM learning.competency_clusters c
JOIN learning.user_profile p ON p.cluster_id = c.id
WHERE p.user_id = sqlc.arg(user_id)
ORDER BY c.code
FOR UPDATE OF p;

-- name: UpdateProfileScore :exec
UPDATE learning.user_profile
SET score = sqlc.arg(score), updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND cluster_id = sqlc.arg(cluster_id);

-- name: CreateProfileHistory :exec
INSERT INTO learning.profile_history (id, user_id, source, source_id, changes)
VALUES (sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(source), sqlc.arg(source_id), sqlc.arg(changes));
