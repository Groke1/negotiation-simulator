-- name: CreateScenario :one
INSERT INTO scenarios.scenarios (
    id,
    organization_id,
    created_by,
    version,
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
    sqlc.arg(version),
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
ORDER BY updated_at DESC, id DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: UpdateScenario :one
UPDATE scenarios.scenarios
SET version = sqlc.arg(version),
    title = sqlc.arg(title),
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
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateScenarioVersion :one
INSERT INTO scenarios.scenario_versions (id, scenario_id, version, snapshot)
VALUES (sqlc.arg(id), sqlc.arg(scenario_id), sqlc.arg(version), sqlc.arg(snapshot))
RETURNING id, scenario_id, version, snapshot, published_at;

-- name: ListScenarioVersions :many
SELECT version, snapshot, published_at
FROM scenarios.scenario_versions
WHERE scenario_id = sqlc.arg(scenario_id)
ORDER BY published_at DESC, version DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: GetScenarioVersion :one
SELECT version, snapshot, published_at
FROM scenarios.scenario_versions
WHERE scenario_id = sqlc.arg(scenario_id)
  AND version = sqlc.arg(version);
