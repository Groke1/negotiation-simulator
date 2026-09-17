-- Scenario = описание переговорного мира.
-- ScenarioVersion = immutable snapshot конкретной опубликованной версии.
--
-- Вложенные части Scenario хранятся в JSONB/массивах, а часто используемые
-- метаданные остаются отдельными колонками.

-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS scenarios;

CREATE TYPE scenarios.scenario_status AS ENUM ('draft', 'published', 'archived');

CREATE TYPE scenarios.scenario_mode AS ENUM ('training', 'arena', 'live');

CREATE TYPE scenarios.scenario_channel AS ENUM ('chat', 'voice', 'meeting');

CREATE TYPE scenarios.scenario_difficulty AS ENUM ('L1', 'L2', 'L3');

CREATE TYPE scenarios.ai_provider AS ENUM ('mock', 'cloud', 'ollama');

-- Общий словарь действий нужен и Constructor, и Event Log.
CREATE TYPE scenarios.semantic_action AS ENUM (
    'ASK_QUESTION',
    'MAKE_PROPOSAL',
    'MAKE_CONCESSION',
    'REQUEST_RECIPROCITY',
    'HOLD_BOUNDARY',
    'CHANGE_TOPIC',
    'PAUSE',
    'SUMMARIZE',
    'CLOSE_NEGOTIATION'
);

CREATE TABLE scenarios.scenarios (
     id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),

     organization_id UUID REFERENCES organizations.organizations(id)
        ON DELETE SET NULL,

     created_by UUID REFERENCES users.users(id) ON DELETE SET NULL,

    -- Текущая версия редактируемого состояния.
    -- При publish эта строка становится version snapshot.
     version                 VARCHAR(32) NOT NULL DEFAULT '1.0',

     title                   VARCHAR(200) NOT NULL,

     status                  scenarios.scenario_status NOT NULL DEFAULT 'draft',

     mode_support            TEXT[] NOT NULL DEFAULT '{}',

    -- {
    --   situation_type,
    --   conflict_type,
    --   description
    -- }
     context                 JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- [{id: learner|partner, label: "..."}]
     roles                   JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- {learner: "...", partner: "..."}
     goals                   JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- {learner: [...], partner: [...]}
     interests               JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Ограничения обеих сторон.
     constraints             JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- [{key, initial, unit?}]
    parameters              JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Закрытая информация сторон.
     private_information     JSONB NOT NULL DEFAULT '{}'::jsonb,

     difficulty              scenarios.scenario_difficulty NOT NULL DEFAULT 'L1',

    -- Названия 6 competency clusters.
    competencies            TEXT[] NOT NULL DEFAULT '{}',

    -- [{semantic_action, cluster, indicator, trigger_condition}]
    observation_rules       JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- [{type: agreement|no_deal, condition: "..."}]
    completion_conditions   JSONB NOT NULL DEFAULT '[]'::jsonb,

    channels                TEXT[] NOT NULL DEFAULT '{}',

    -- Детерминированный граф обычной тренировки:
    -- nodes -> actions -> next_node_id.
    nodes                   JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- revision увеличивается при каждом изменении Scenario.
    -- validated_revision хранит revision, которая последний раз успешно
    -- прошла POST /scenarios/{id}/validate.
     revision                BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0),
     validated_revision      BIGINT NULL,
     validation_result       JSONB NULL,

     created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

     CONSTRAINT chk_scenario_validated_revision
         CHECK (
             validated_revision IS NULL
                 OR (
                 validated_revision > 0
                     AND validated_revision <= revision
                 )
             )
);

CREATE INDEX idx_scenarios_created_by
    ON scenarios.scenarios(created_by);

CREATE INDEX idx_scenarios_organization_id
    ON scenarios.scenarios(organization_id);

CREATE INDEX idx_scenarios_created_by_status
    ON scenarios.scenarios(created_by, status);

CREATE INDEX idx_scenarios_organization_status
    ON scenarios.scenarios(organization_id, status);


-- Immutable snapshots.
CREATE TABLE scenarios.scenario_versions (
     id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),

     scenario_id UUID NOT NULL REFERENCES scenarios.scenarios(id)
        ON DELETE RESTRICT,
     version         VARCHAR(32) NOT NULL,

    -- Полная копия Scenario на момент publish.
     snapshot        JSONB NOT NULL,

     published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

     UNIQUE (scenario_id, version),

    -- Нужен для composite FK из training.sessions и assignments:
    -- version_id должен действительно принадлежать указанному scenario_id.
     UNIQUE (id, scenario_id)
);

CREATE INDEX idx_scenario_versions_scenario_id
    ON scenarios.scenario_versions(scenario_id);

CREATE INDEX idx_scenario_versions_published_at
    ON scenarios.scenario_versions(scenario_id, published_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS scenarios.idx_scenario_versions_published_at;
DROP INDEX IF EXISTS scenarios.idx_scenario_versions_scenario_id;
DROP TABLE IF EXISTS scenarios.scenario_versions;

DROP INDEX IF EXISTS scenarios.idx_scenarios_organization_status;
DROP INDEX IF EXISTS scenarios.idx_scenarios_created_by_status;
DROP INDEX IF EXISTS scenarios.idx_scenarios_organization_id;
DROP INDEX IF EXISTS scenarios.idx_scenarios_created_by;
DROP TABLE IF EXISTS scenarios.scenarios;

DROP TYPE IF EXISTS scenarios.semantic_action;
DROP TYPE IF EXISTS scenarios.ai_provider;
DROP TYPE IF EXISTS scenarios.scenario_difficulty;
DROP TYPE IF EXISTS scenarios.scenario_channel;
DROP TYPE IF EXISTS scenarios.scenario_mode;
DROP TYPE IF EXISTS scenarios.scenario_status;

DROP SCHEMA IF EXISTS scenarios;

-- +goose StatementEnd