-- name: CreateArenaAssignment :one
INSERT INTO learning.arena_assignments (id, scenario_id, created_by, participants, channel, scheduled_at)
VALUES (sqlc.arg(id), sqlc.arg(scenario_id), sqlc.arg(created_by), sqlc.arg(participants), sqlc.arg(channel), sqlc.arg(scheduled_at))
RETURNING *;

-- name: GetArenaAssignment :one
SELECT aa.*, sc.title AS scenario_title
FROM learning.arena_assignments aa
JOIN scenarios.scenarios sc ON sc.id = aa.scenario_id
WHERE aa.id = sqlc.arg(id);

-- name: ListArenaAssignments :many
SELECT aa.*, sc.title AS scenario_title
FROM learning.arena_assignments aa
JOIN scenarios.scenarios sc ON sc.id = aa.scenario_id
WHERE aa.created_by = sqlc.arg(curator_id)
ORDER BY aa.scheduled_at DESC, aa.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: ListUserArenaAssignments :many
SELECT aa.*, sc.title AS scenario_title
FROM learning.arena_assignments aa
JOIN scenarios.scenarios sc ON sc.id = aa.scenario_id
WHERE aa.participants @> sqlc.arg(participant_filter)::jsonb
ORDER BY aa.scheduled_at DESC, aa.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CreateArenaSession :one
INSERT INTO learning.arena_sessions (id, assignment_id)
VALUES (sqlc.arg(id), sqlc.arg(assignment_id))
ON CONFLICT (assignment_id) DO UPDATE SET assignment_id = EXCLUDED.assignment_id
RETURNING *;

-- name: SetArenaAssignmentSession :one
UPDATE learning.arena_assignments
SET session_id = sqlc.arg(session_id), status = sqlc.arg(status)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetArenaSession :one
SELECT s.id, s.assignment_id, s.status, s.joined, s.started_at, s.completed_at,
       aa.scenario_id, sc.title AS scenario_title, aa.participants, aa.channel, aa.scheduled_at, aa.created_by
FROM learning.arena_sessions s
JOIN learning.arena_assignments aa ON aa.id = s.assignment_id
JOIN scenarios.scenarios sc ON sc.id = aa.scenario_id
WHERE s.id = sqlc.arg(id);

-- name: UpdateArenaSessionJoin :one
UPDATE learning.arena_sessions
SET joined = sqlc.arg(joined), status = sqlc.arg(status), started_at = COALESCE(started_at, sqlc.narg(started_at))
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CompleteArenaSession :one
UPDATE learning.arena_sessions
SET status = 'completed', completed_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CompleteArenaAssignment :exec
UPDATE learning.arena_assignments
SET status = 'completed'
WHERE id = sqlc.arg(id);

-- name: CreateArenaMessage :one
INSERT INTO learning.arena_messages (id, session_id, sender_id, content, semantic_action_code, semantic_annotations)
VALUES (sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(sender_id), sqlc.arg(content), sqlc.narg(semantic_action_code), sqlc.narg(semantic_annotations))
RETURNING *;

-- name: ListArenaMessages :many
SELECT *
FROM learning.arena_messages
WHERE session_id = sqlc.arg(session_id)
ORDER BY created_at, id;

-- name: CreateArenaFeedback :one
INSERT INTO learning.arena_feedback (id, session_id, participant_id, curator_id, comment, ratings, profile_change)
VALUES (sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(participant_id), sqlc.arg(curator_id), sqlc.arg(comment), sqlc.arg(ratings), sqlc.narg(profile_change))
RETURNING *;

-- name: ListArenaFeedback :many
SELECT *
FROM learning.arena_feedback
WHERE session_id = sqlc.arg(session_id)
ORDER BY created_at, id;

-- name: ListUserArenaHistory :many
SELECT s.id AS arena_session_id, aa.id AS arena_assignment_id, aa.scenario_id, sc.title AS scenario_title,
       aa.scheduled_at, s.started_at, s.completed_at, s.status,
       EXISTS (SELECT 1 FROM learning.arena_feedback af WHERE af.session_id = s.id AND af.participant_id = sqlc.arg(user_id)) AS has_feedback
FROM learning.arena_sessions s
JOIN learning.arena_assignments aa ON aa.id = s.assignment_id
JOIN scenarios.scenarios sc ON sc.id = aa.scenario_id
WHERE aa.participants @> sqlc.arg(participant_filter)::jsonb
ORDER BY aa.scheduled_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);
