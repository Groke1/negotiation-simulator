-- name: ListCompetencyClusters :many
SELECT id, code, name, description
FROM learning.competency_clusters
ORDER BY code;

-- name: ListProfileForUpdate :many
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
ORDER BY c.code
FOR UPDATE OF p;

-- name: UpdateProfileState :exec
UPDATE learning.user_profile
SET score = sqlc.arg(score), evidence_mass = sqlc.arg(evidence_mass), confidence = sqlc.arg(confidence), updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
  AND cluster_id = sqlc.arg(cluster_id);

-- name: CreateProfileHistory :exec
INSERT INTO learning.profile_history (id, user_id, source, source_id, changes)
VALUES (sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(source), sqlc.arg(source_id), sqlc.arg(changes));

-- name: ListSemanticActions :many
SELECT code, name, description
FROM learning.semantic_actions
ORDER BY code;
