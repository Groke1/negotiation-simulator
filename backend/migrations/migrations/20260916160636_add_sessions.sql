-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS training;

CREATE TYPE training.session_status AS ENUM (
    'IN_PROGRESS',
    'READY_TO_COMPLETE',
    'COMPLETED',
    'ABANDONED'
);

CREATE TYPE training.channel AS ENUM (
    'chat',
    'voice'
);

CREATE TABLE training.sessions (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

   user_id UUID NOT NULL
       REFERENCES users.users(id),

   scenario_id UUID NOT NULL
   REFERENCES scenarios.scenarios(id),

    -- Копия состояния сценария именно для этой сессии.
    -- Для preview особенно важна.
   scenario_snapshot JSONB NOT NULL,

   channel training.channel NOT NULL,

   is_preview BOOLEAN NOT NULL DEFAULT FALSE,

   status training.session_status NOT NULL DEFAULT 'IN_PROGRESS',

   current_node_id VARCHAR(255) NULL,

   outcome JSONB NULL,
   analysis JSONB NULL,

   created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   completed_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_training_sessions_user_id
    ON training.sessions(user_id);

CREATE INDEX idx_training_sessions_scenario_id
    ON training.sessions(scenario_id);

CREATE INDEX idx_training_sessions_user_created_at
    ON training.sessions(user_id, created_at DESC);


CREATE TABLE training.session_events (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

     session_id UUID NOT NULL
         REFERENCES training.sessions(id)
             ON DELETE CASCADE,

    -- idempotency key от frontend
     client_event_id UUID NOT NULL,

     sequence_number INTEGER NOT NULL,

     node_id VARCHAR(255) NOT NULL,
     action_id VARCHAR(255) NOT NULL,

     semantic_action VARCHAR(100) NOT NULL,

     next_node_id VARCHAR(255) NULL,

    -- Если потом понадобятся дополнительные данные события.
     payload JSONB NOT NULL DEFAULT '{}'::jsonb,

     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

     UNIQUE (session_id, client_event_id),
     UNIQUE (session_id, sequence_number)
);

CREATE INDEX idx_session_events_session_id
    ON training.session_events(session_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS training.session_events;
DROP TABLE IF EXISTS training.sessions;

DROP TYPE IF EXISTS training.channel;
DROP TYPE IF EXISTS training.session_status;

DROP SCHEMA IF EXISTS training;

-- +goose StatementEnd
