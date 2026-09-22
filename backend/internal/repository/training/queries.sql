-- name: CreateTrainingSession :one
INSERT INTO learning.training_sessions (
    id, user_id, scenario_id, scenario_role_id, mode, channel, status, assignment_id, snapshot, state
)
VALUES (
    sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(scenario_id), sqlc.arg(scenario_role_id),
    sqlc.arg(mode), sqlc.arg(channel), sqlc.arg(status), sqlc.narg(assignment_id),
    sqlc.arg(snapshot), sqlc.arg(state)
)
RETURNING *;

-- name: GetTrainingSession :one
SELECT s.*, sc.title AS scenario_title
FROM learning.training_sessions s
JOIN scenarios.scenarios sc ON sc.id = s.scenario_id
WHERE s.id = sqlc.arg(id);

-- name: ListTrainingSessions :many
SELECT s.*, sc.title AS scenario_title
FROM learning.training_sessions s
JOIN scenarios.scenarios sc ON sc.id = s.scenario_id
WHERE s.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(scenario_id)::uuid IS NULL OR s.scenario_id = sqlc.narg(scenario_id))
  AND (sqlc.narg(status)::text IS NULL OR s.status = sqlc.narg(status))
  AND (sqlc.narg(mode)::text IS NULL OR s.mode = sqlc.narg(mode))
  AND (sqlc.narg(channel)::text IS NULL OR s.channel = sqlc.narg(channel))
ORDER BY s.created_at DESC, s.id DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: UpdateTrainingState :one
UPDATE learning.training_sessions
SET state = sqlc.arg(state), status = sqlc.arg(status)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CompleteTrainingSession :one
UPDATE learning.training_sessions
SET status = 'completed', score = sqlc.arg(score), analysis = sqlc.arg(analysis), completed_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateTrainingMessage :one
INSERT INTO learning.training_messages (
    id, session_id, author, content, semantic_action_code, semantic_annotations, expected_action_code, classification
)
VALUES (
    sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(author), sqlc.arg(content),
    sqlc.narg(semantic_action_code), sqlc.narg(semantic_annotations), sqlc.narg(expected_action_code), sqlc.narg(classification)
)
RETURNING *;

-- name: ListTrainingMessages :many
SELECT *
FROM learning.training_messages
WHERE session_id = sqlc.arg(session_id)
ORDER BY created_at, id;

-- name: GetTrainingMessage :one
SELECT *
FROM learning.training_messages
WHERE id = sqlc.arg(id)
  AND session_id = sqlc.arg(session_id);

-- name: UpdateTrainingMessageClassification :one
UPDATE learning.training_messages
SET classification = sqlc.arg(classification)
WHERE id = sqlc.arg(id)
  AND session_id = sqlc.arg(session_id)
RETURNING *;
