-- +goose Up
-- +goose StatementBegin

CREATE TABLE learning.training_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    scenario_id UUID NOT NULL REFERENCES scenarios.scenarios(id),
    scenario_role_id TEXT NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('deterministic','virtual_partner')),
    channel TEXT NOT NULL CHECK (channel IN ('chat','email','voice_message')),
    status TEXT NOT NULL CHECK (status IN ('in_progress','completed','abandoned')),
    assignment_id UUID NULL,
    snapshot JSONB NOT NULL,
    state JSONB NOT NULL DEFAULT '{}'::jsonb,
    analysis JSONB NULL,
    score DOUBLE PRECISION NULL CHECK (score >= 0 AND score <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_learning_training_sessions_user ON learning.training_sessions(user_id, created_at DESC);
CREATE INDEX idx_learning_training_sessions_scenario ON learning.training_sessions(scenario_id);

CREATE TABLE learning.training_messages (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES learning.training_sessions(id) ON DELETE CASCADE,
    author TEXT NOT NULL CHECK (author IN ('learner','virtual_partner')),
    content JSONB NOT NULL,
    semantic_action_code TEXT NULL,
    semantic_annotations JSONB NULL,
    expected_action_code TEXT NULL,
    classification JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_learning_training_messages_session ON learning.training_messages(session_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS learning.idx_learning_training_messages_session;
DROP TABLE IF EXISTS learning.training_messages;
DROP INDEX IF EXISTS learning.idx_learning_training_sessions_scenario;
DROP INDEX IF EXISTS learning.idx_learning_training_sessions_user;
DROP TABLE IF EXISTS learning.training_sessions;

-- +goose StatementEnd
