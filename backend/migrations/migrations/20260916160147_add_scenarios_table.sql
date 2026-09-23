-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS scenarios;

CREATE TYPE scenarios.scenario_status AS ENUM ('draft', 'published', 'archived');

CREATE TYPE scenarios.scenario_difficulty AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE scenarios.scenarios (
     id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),

     organization_id UUID REFERENCES organizations.organizations(id)
        ON DELETE SET NULL,

     created_by UUID REFERENCES users.users(id) ON DELETE SET NULL,

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

     difficulty              scenarios.scenario_difficulty NOT NULL DEFAULT 'easy',

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
     published_at            TIMESTAMPTZ NULL,

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

CREATE TABLE scenarios.scenario_favorites (
     user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
     scenario_id UUID NOT NULL REFERENCES scenarios.scenarios(id) ON DELETE CASCADE,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (user_id, scenario_id)
);

CREATE INDEX idx_scenario_favorites_scenario_id
    ON scenarios.scenario_favorites(scenario_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS scenarios.idx_scenario_favorites_scenario_id;
DROP TABLE IF EXISTS scenarios.scenario_favorites;

DROP INDEX IF EXISTS scenarios.idx_scenarios_organization_status;
DROP INDEX IF EXISTS scenarios.idx_scenarios_created_by_status;
DROP INDEX IF EXISTS scenarios.idx_scenarios_organization_id;
DROP INDEX IF EXISTS scenarios.idx_scenarios_created_by;
DROP TABLE IF EXISTS scenarios.scenarios;

DROP TYPE IF EXISTS scenarios.scenario_difficulty;
DROP TYPE IF EXISTS scenarios.scenario_status;

DROP SCHEMA IF EXISTS scenarios;

-- +goose StatementEnd
