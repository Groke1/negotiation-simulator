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

-- name: FindPendingAssignmentForTraining :one
SELECT id
FROM learning.assignments
WHERE learner_id = sqlc.arg(learner_id)
  AND scenario_id = sqlc.arg(scenario_id)
  AND status = 'pending'
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: PendingAssignmentMatchesTraining :one
SELECT EXISTS (
    SELECT 1
    FROM learning.assignments
    WHERE id = sqlc.arg(assignment_id)
      AND learner_id = sqlc.arg(learner_id)
      AND scenario_id = sqlc.arg(scenario_id)
      AND status = 'pending'
);

-- name: GetTrainingSession :one
SELECT s.*, sc.title AS scenario_title
FROM learning.training_sessions s
JOIN scenarios.scenarios sc ON sc.id = s.scenario_id
WHERE s.id = sqlc.arg(id);

-- name: ListTrainingSessions :many
SELECT s.*, sc.title AS scenario_title
FROM learning.training_sessions s
JOIN scenarios.scenarios sc ON sc.id = s.scenario_id
JOIN users.users learner ON learner.id = s.user_id
LEFT JOIN users.users curator ON curator.id = sqlc.narg(curator_id)
WHERE (sqlc.narg(user_id)::uuid IS NULL OR s.user_id = sqlc.narg(user_id))
  AND (sqlc.narg(curator_id)::uuid IS NULL OR (curator.organization_id IS NOT NULL AND learner.organization_id = curator.organization_id))
  AND (sqlc.narg(scenario_id)::uuid IS NULL OR s.scenario_id = sqlc.narg(scenario_id))
  AND (sqlc.narg(status)::text IS NULL OR s.status = sqlc.narg(status))
  AND (sqlc.narg(mode)::text IS NULL OR s.mode = sqlc.narg(mode))
  AND (sqlc.narg(channel)::text IS NULL OR s.channel = sqlc.narg(channel))
ORDER BY s.created_at DESC, s.id DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: TrainingSessionInCuratorOrganization :one
SELECT EXISTS (
    SELECT 1
    FROM learning.training_sessions s
    JOIN users.users learner ON learner.id = s.user_id
    JOIN users.users curator ON curator.id = sqlc.arg(curator_id)
    WHERE s.id = sqlc.arg(session_id)
      AND curator.organization_id IS NOT NULL
      AND learner.organization_id = curator.organization_id
);

-- name: UpdateTrainingState :one
UPDATE learning.training_sessions
SET state = sqlc.arg(state),
    status = sqlc.arg(status),
    completed_at = CASE
        WHEN sqlc.arg(status) = 'completed' THEN COALESCE(completed_at, NOW())
        ELSE completed_at
    END
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CompleteTrainingSession :one
UPDATE learning.training_sessions
SET status = 'completed', score = sqlc.arg(score), analysis = sqlc.arg(analysis), completed_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetTrainingAnalysis :one
UPDATE learning.training_sessions
SET analysis = sqlc.arg(analysis)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CompleteLinkedAssignment :exec
UPDATE learning.assignments
SET status = 'completed', completed_at = NOW()
WHERE learning.assignments.id = (
    SELECT assignment_id
    FROM learning.training_sessions
    WHERE learning.training_sessions.id = sqlc.arg(session_id)
)
  AND status = 'pending';

-- name: CreateTrainingMessage :one
INSERT INTO learning.training_messages (
    id, session_id, author, content, semantic_action_code, semantic_annotations, expected_action_code, classification, created_at
)
VALUES (
    sqlc.arg(id), sqlc.arg(session_id), sqlc.arg(author), sqlc.arg(content),
    sqlc.narg(semantic_action_code), sqlc.narg(semantic_annotations), sqlc.narg(expected_action_code), sqlc.narg(classification),
    clock_timestamp()
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
