-- name: CreateArenaAssignment :one
INSERT INTO arena.assignments (
    id, organization_id, scenario_id, scenario_version_id, scenario_title, scenario_version, created_by, scheduled_at
) VALUES (
    sqlc.arg(id), sqlc.arg(organization_id), sqlc.arg(scenario_id), sqlc.arg(scenario_version_id),
    sqlc.arg(scenario_title), sqlc.arg(scenario_version), sqlc.arg(created_by), sqlc.arg(scheduled_at)
)
RETURNING *;

-- name: CreateArenaAssignmentParticipant :exec
INSERT INTO arena.assignment_participants (assignment_id, user_id, role)
VALUES (sqlc.arg(assignment_id), sqlc.arg(user_id), sqlc.arg(role));

-- name: GetArenaAssignment :one
SELECT a.*, s.id AS session_id
FROM arena.assignments a
LEFT JOIN arena.sessions s ON s.assignment_id = a.id
WHERE a.id = sqlc.arg(id);

-- name: ListArenaAssignmentParticipants :many
SELECT * FROM arena.assignment_participants WHERE assignment_id = sqlc.arg(assignment_id) ORDER BY role;

-- name: ListUserArenaAssignments :many
SELECT a.*, s.id AS session_id
FROM arena.assignments a
JOIN arena.assignment_participants ap ON ap.assignment_id = a.id
LEFT JOIN arena.sessions s ON s.assignment_id = a.id
WHERE ap.user_id = sqlc.arg(user_id)
  AND a.status IN ('scheduled', 'ready', 'in_progress')
ORDER BY a.scheduled_at ASC;

-- name: ListOrganizationArenaAssignments :many
SELECT a.*, s.id AS session_id
FROM arena.assignments a
LEFT JOIN arena.sessions s ON s.assignment_id = a.id
WHERE a.organization_id = sqlc.arg(organization_id)
ORDER BY a.scheduled_at DESC;

-- name: CreateArenaSession :one
INSERT INTO arena.sessions (id, assignment_id)
VALUES (sqlc.arg(id), sqlc.arg(assignment_id))
ON CONFLICT (assignment_id) DO UPDATE SET assignment_id = EXCLUDED.assignment_id
RETURNING *;

-- name: GetArenaSession :one
SELECT s.*, a.scenario_id, a.scenario_title, a.scenario_version, a.scheduled_at
FROM arena.sessions s
JOIN arena.assignments a ON a.id = s.assignment_id
WHERE s.id = sqlc.arg(id);

-- name: AddArenaSessionParticipant :exec
INSERT INTO arena.session_participants (session_id, user_id, role)
VALUES (sqlc.arg(session_id), sqlc.arg(user_id), sqlc.arg(role))
ON CONFLICT (session_id, user_id) DO NOTHING;

-- name: ListArenaSessionParticipants :many
SELECT * FROM arena.session_participants WHERE session_id = sqlc.arg(session_id) ORDER BY role;

-- name: CountArenaSessionParticipants :one
SELECT count(*)::int FROM arena.session_participants WHERE session_id = sqlc.arg(session_id);

-- name: SetArenaSessionInProgress :one
UPDATE arena.sessions
SET status = 'in_progress', started_at = COALESCE(started_at, NOW()), updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'waiting_for_participants'
RETURNING *;

-- name: CompleteArenaSession :one
UPDATE arena.sessions
SET status = 'completed', completed_at = NOW(), updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'in_progress'
RETURNING *;

-- name: SetArenaAssignmentStatus :exec
UPDATE arena.assignments SET status = sqlc.arg(status), updated_at = NOW() WHERE id = sqlc.arg(id);

-- name: CreateArenaMessage :one
INSERT INTO arena.messages (id, session_id, sender_id, text)
VALUES (sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(sender_id), sqlc.arg(text))
ON CONFLICT (id) DO UPDATE SET id = EXCLUDED.id
RETURNING *;

-- name: ListArenaMessages :many
SELECT * FROM arena.messages WHERE session_id = sqlc.arg(session_id) ORDER BY created_at ASC;

-- name: CreateArenaFeedback :one
INSERT INTO arena.feedback (id, session_id, participant_id, expert_id, comment, ratings)
VALUES (sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(participant_id), sqlc.arg(expert_id), sqlc.arg(comment), sqlc.arg(ratings))
RETURNING *;

-- name: ListArenaFeedback :many
SELECT * FROM arena.feedback WHERE session_id = sqlc.arg(session_id) ORDER BY created_at ASC;

-- name: ListUserArenaHistory :many
SELECT s.*, a.scenario_id, a.scenario_title, a.scenario_version, a.scheduled_at,
       EXISTS (SELECT 1 FROM arena.feedback f WHERE f.session_id = s.id AND f.participant_id = sqlc.arg(user_id)) AS has_feedback
FROM arena.sessions s
JOIN arena.assignments a ON a.id = s.assignment_id
JOIN arena.assignment_participants ap ON ap.assignment_id = a.id
WHERE ap.user_id = sqlc.arg(user_id) AND s.status IN ('completed', 'reviewed')
ORDER BY s.completed_at DESC;
