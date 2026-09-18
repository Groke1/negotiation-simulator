-- name: CreateTrainingSession :one
INSERT INTO training.sessions (
    id, user_id, scenario_id, scenario_version_id, scenario_snapshot,
    mode, channel, is_preview, status, current_node_id
) VALUES (
    sqlc.arg(id), sqlc.arg(user_id), sqlc.arg(scenario_id), sqlc.narg(scenario_version_id), sqlc.arg(scenario_snapshot),
    sqlc.arg(mode), sqlc.arg(channel), sqlc.arg(is_preview), 'IN_PROGRESS', sqlc.narg(current_node_id)
)
RETURNING *;

-- name: GetTrainingSessionByID :one
SELECT * FROM training.sessions WHERE id = sqlc.arg(id);

-- name: UpdateTrainingSessionState :one
UPDATE training.sessions
SET current_node_id = sqlc.narg(current_node_id), status = sqlc.arg(status), updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CompleteTrainingSession :one
UPDATE training.sessions
SET status = 'COMPLETED', outcome = sqlc.arg(outcome), analysis = sqlc.arg(analysis), completed_at = NOW(), updated_at = NOW()
WHERE id = sqlc.arg(id) AND status IN ('IN_PROGRESS', 'READY_TO_COMPLETE')
RETURNING *;

-- name: CreateTrainingEvent :one
INSERT INTO training.session_events (
    id, session_id, client_event_id, sequence_number, node_id, action_id, semantic_action, next_node_id
) VALUES (
    sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(client_event_id), sqlc.arg(sequence_number),
    sqlc.arg(node_id), sqlc.arg(action_id), sqlc.arg(semantic_action), sqlc.narg(next_node_id)
)
ON CONFLICT (session_id, client_event_id) DO UPDATE SET client_event_id = EXCLUDED.client_event_id
RETURNING *;

-- name: CountTrainingEvents :one
SELECT count(*)::int FROM training.session_events WHERE session_id = sqlc.arg(session_id);

-- name: ListTrainingEvents :many
SELECT * FROM training.session_events WHERE session_id = sqlc.arg(session_id) ORDER BY sequence_number ASC;

-- name: CreateTrainingMessage :one
INSERT INTO training.session_messages (id, session_id, author, text)
VALUES (sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(author), sqlc.arg(text))
ON CONFLICT (id) DO UPDATE SET id = EXCLUDED.id
RETURNING *;

-- name: ListTrainingMessages :many
SELECT * FROM training.session_messages WHERE session_id = sqlc.arg(session_id) ORDER BY created_at ASC;
