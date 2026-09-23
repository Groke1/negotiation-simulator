-- name: CreateScenario :one
INSERT INTO scenarios.scenarios (
    id,
    organization_id,
    created_by,
    title,
    mode_support,
    context,
    roles,
    goals,
    interests,
    constraints,
    parameters,
    private_information,
    difficulty,
    competencies,
    observation_rules,
    completion_conditions,
    channels,
    nodes
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(organization_id),
    sqlc.arg(created_by),
    sqlc.arg(title),
    sqlc.arg(mode_support),
    sqlc.arg(context),
    sqlc.arg(roles),
    sqlc.arg(goals),
    sqlc.arg(interests),
    sqlc.arg(constraints),
    sqlc.arg(parameters),
    sqlc.arg(private_information),
    sqlc.arg(difficulty),
    sqlc.arg(competencies),
    sqlc.arg(observation_rules),
    sqlc.arg(completion_conditions),
    sqlc.arg(channels),
    sqlc.arg(nodes)
)
RETURNING *;

-- name: GetScenarioByID :one
SELECT *
FROM scenarios.scenarios
WHERE id = sqlc.arg(id);

-- name: ListScenariosByOrganization :many
SELECT *
FROM scenarios.scenarios
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.arg(status)::text = '' OR status = sqlc.arg(status)::scenarios.scenario_status)
  AND (sqlc.arg(difficulty)::text = '' OR difficulty = sqlc.arg(difficulty)::scenarios.scenario_difficulty)
  AND (sqlc.arg(tag)::text = '' OR context->'tags' ? sqlc.arg(tag)::text)
  AND (sqlc.arg(runtime)::text = '' OR sqlc.arg(runtime)::text = ANY(mode_support))
  AND (sqlc.arg(channel)::text = '' OR sqlc.arg(channel)::text = ANY(channels))
  AND (NOT sqlc.arg(owned_by_me)::boolean OR created_by = sqlc.arg(user_id))
  AND (
      sqlc.narg(favorite)::boolean IS NULL
      OR EXISTS (
          SELECT 1 FROM scenarios.scenario_favorites sf
          WHERE sf.scenario_id = scenarios.scenarios.id
            AND sf.user_id = sqlc.arg(user_id)
      ) = sqlc.narg(favorite)::boolean
  )
  AND (
      sqlc.narg(assigned_by_curator)::boolean IS NULL
      OR EXISTS (
          SELECT 1 FROM learning.assignments a
          WHERE a.scenario_id = scenarios.scenarios.id
            AND a.learner_id = sqlc.arg(user_id)
            AND a.status = 'pending'
      ) = sqlc.narg(assigned_by_curator)::boolean
  )
  AND (
      sqlc.narg(completed)::boolean IS NULL
      OR EXISTS (
          SELECT 1 FROM learning.training_sessions ts
          WHERE ts.scenario_id = scenarios.scenarios.id
            AND ts.user_id = sqlc.arg(user_id)
            AND ts.status = 'completed'
      ) = sqlc.narg(completed)::boolean
  )
ORDER BY updated_at DESC, id DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: UpdateScenario :one
UPDATE scenarios.scenarios
SET title = sqlc.arg(title),
    mode_support = sqlc.arg(mode_support),
    context = sqlc.arg(context),
    roles = sqlc.arg(roles),
    goals = sqlc.arg(goals),
    interests = sqlc.arg(interests),
    constraints = sqlc.arg(constraints),
    parameters = sqlc.arg(parameters),
    private_information = sqlc.arg(private_information),
    difficulty = sqlc.arg(difficulty),
    competencies = sqlc.arg(competencies),
    observation_rules = sqlc.arg(observation_rules),
    completion_conditions = sqlc.arg(completion_conditions),
    channels = sqlc.arg(channels),
    nodes = sqlc.arg(nodes),
    revision = revision + 1,
    validated_revision = NULL,
    validation_result = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SaveScenarioValidation :one
UPDATE scenarios.scenarios
SET validated_revision = revision,
    validation_result = sqlc.arg(validation_result),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: SetScenarioStatus :one
UPDATE scenarios.scenarios
SET status = sqlc.arg(status),
    published_at = CASE
        WHEN sqlc.arg(status) = 'published'::scenarios.scenario_status THEN COALESCE(published_at, NOW())
        ELSE published_at
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AddScenarioFavorite :exec
INSERT INTO scenarios.scenario_favorites (user_id, scenario_id)
VALUES (sqlc.arg(user_id), sqlc.arg(scenario_id))
ON CONFLICT DO NOTHING;

-- name: RemoveScenarioFavorite :exec
DELETE FROM scenarios.scenario_favorites
WHERE user_id = sqlc.arg(user_id)
  AND scenario_id = sqlc.arg(scenario_id);

-- name: GetScenarioPersonalStats :one
WITH attempts AS (
    SELECT
        COUNT(*)::integer AS attempts_count,
        COUNT(*) FILTER (WHERE status = 'completed')::integer AS completed_count
    FROM learning.training_sessions
    WHERE user_id = sqlc.arg(user_id)
      AND scenario_id = sqlc.arg(scenario_id)
),
best AS (
    SELECT id, score::integer AS score, completed_at
    FROM learning.training_sessions
    WHERE user_id = sqlc.arg(user_id)
      AND scenario_id = sqlc.arg(scenario_id)
      AND status = 'completed'
      AND score IS NOT NULL
      AND completed_at IS NOT NULL
    ORDER BY score DESC, completed_at DESC
    LIMIT 1
)
SELECT
    EXISTS (
        SELECT 1 FROM scenarios.scenario_favorites sf
        WHERE sf.user_id = sqlc.arg(user_id)
          AND sf.scenario_id = sqlc.arg(scenario_id)
    ) AS is_favorite,
    EXISTS (
        SELECT 1 FROM learning.assignments a
        WHERE a.learner_id = sqlc.arg(user_id)
          AND a.scenario_id = sqlc.arg(scenario_id)
          AND a.status = 'pending'
    ) AS is_assigned_by_curator,
    COALESCE((SELECT completed_count > 0 FROM attempts), false)::boolean AS is_completed,
    COALESCE((SELECT attempts_count FROM attempts), 0)::integer AS attempts_count,
    (SELECT id FROM best) AS best_session_id,
    COALESCE((SELECT score FROM best), 0)::integer AS best_score,
    (SELECT completed_at FROM best) AS best_completed_at;
