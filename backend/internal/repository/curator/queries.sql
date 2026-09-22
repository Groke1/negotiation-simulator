-- name: CreateAssignment :one
INSERT INTO learning.assignments (id, curator_id, learner_id, scenario_id)
VALUES (sqlc.arg(id), sqlc.arg(curator_id), sqlc.arg(learner_id), sqlc.arg(scenario_id))
RETURNING id, curator_id, learner_id, scenario_id, status, created_at, completed_at;

-- name: ListAssignments :many
SELECT a.id, a.curator_id, a.learner_id, a.scenario_id, sc.title AS scenario_title, a.status, a.created_at, a.completed_at
FROM learning.assignments a
JOIN scenarios.scenarios sc ON sc.id = a.scenario_id
WHERE a.curator_id = sqlc.arg(curator_id)
  AND (sqlc.narg(status)::text IS NULL OR a.status = sqlc.narg(status))
ORDER BY a.created_at DESC, a.id DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: ListCuratorLearners :many
SELECT u.id, u.username, u.email, u.created_at,
       COALESCE(COUNT(ts.id), 0)::integer AS sessions_count,
       MAX(ts.completed_at)::timestamptz AS last_session_at,
       COALESCE(AVG(ts.score)::integer, -1)::int4 AS average_score
FROM users.users curator
JOIN users.users u ON u.organization_id = curator.organization_id
JOIN users.user_roles ur ON ur.user_id = u.id AND ur.role = 'learner'
LEFT JOIN learning.training_sessions ts ON ts.user_id = u.id AND ts.status = 'completed'
WHERE curator.id = sqlc.arg(curator_id)
  AND u.id <> curator.id
GROUP BY u.id, u.username, u.email, u.created_at
ORDER BY u.username
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: ListTeamTrainingSessions :many
SELECT 'training'::text AS type, ts.id AS session_id, ARRAY[ts.user_id]::uuid[] AS user_ids,
       ts.scenario_id, sc.title AS scenario_title, ts.status, COALESCE(ts.score::integer, -1)::int4 AS score,
       ts.created_at AS started_at, ts.completed_at
FROM learning.training_sessions ts
JOIN scenarios.scenarios sc ON sc.id = ts.scenario_id
JOIN users.users learner ON learner.id = ts.user_id
JOIN users.users curator ON curator.id = sqlc.arg(curator_id)
WHERE learner.organization_id = curator.organization_id
ORDER BY ts.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: TeamAnalytics :one
SELECT
    COUNT(DISTINCT u.id)::integer AS users_count,
    COUNT(DISTINCT ts.id)::integer AS training_sessions_count,
    COUNT(DISTINCT ars.id)::integer AS arena_sessions_count,
    COUNT(DISTINCT a.id) FILTER (WHERE a.status = 'pending')::integer AS pending_assignments_count
FROM users.users curator
LEFT JOIN users.users u ON u.organization_id = curator.organization_id AND u.id <> curator.id
LEFT JOIN learning.training_sessions ts ON ts.user_id = u.id
LEFT JOIN learning.arena_feedback af ON af.participant_id = u.id
LEFT JOIN learning.arena_sessions ars ON ars.id = af.session_id
LEFT JOIN learning.assignments a ON a.curator_id = curator.id
WHERE curator.id = sqlc.arg(curator_id);
