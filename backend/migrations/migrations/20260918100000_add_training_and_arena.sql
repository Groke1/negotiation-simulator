-- +goose Up
-- +goose StatementBegin

CREATE TABLE training.session_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES training.sessions(id) ON DELETE CASCADE,
    author TEXT NOT NULL CHECK (author IN ('learner', 'virtual_partner')),
    text TEXT NOT NULL CHECK (char_length(text) BETWEEN 1 AND 5000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_training_session_messages_session_created_at
    ON training.session_messages(session_id, created_at);

CREATE SCHEMA IF NOT EXISTS arena;

CREATE TYPE arena.assignment_status AS ENUM ('scheduled', 'ready', 'in_progress', 'completed', 'cancelled');
CREATE TYPE arena.session_status AS ENUM ('waiting_for_participants', 'in_progress', 'completed', 'reviewed');
CREATE TYPE arena.participant_role AS ENUM ('learner', 'partner');

CREATE TABLE arena.assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations.organizations(id),
    scenario_id UUID NOT NULL REFERENCES scenarios.scenarios(id),
    scenario_version_id UUID NOT NULL REFERENCES scenarios.scenario_versions(id),
    scenario_title TEXT NOT NULL,
    scenario_version TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users.users(id),
    scheduled_at TIMESTAMPTZ NOT NULL,
    status arena.assignment_status NOT NULL DEFAULT 'scheduled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE arena.assignment_participants (
    assignment_id UUID NOT NULL REFERENCES arena.assignments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users.users(id),
    role arena.participant_role NOT NULL,
    PRIMARY KEY (assignment_id, user_id),
    UNIQUE (assignment_id, role)
);

CREATE TABLE arena.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL UNIQUE REFERENCES arena.assignments(id),
    status arena.session_status NOT NULL DEFAULT 'waiting_for_participants',
    channel TEXT NOT NULL DEFAULT 'chat' CHECK (channel = 'chat'),
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE arena.session_participants (
    session_id UUID NOT NULL REFERENCES arena.sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users.users(id),
    role arena.participant_role NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (session_id, user_id)
);

CREATE TABLE arena.messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES arena.sessions(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users.users(id),
    text TEXT NOT NULL CHECK (char_length(text) BETWEEN 1 AND 5000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_arena_messages_session_created_at ON arena.messages(session_id, created_at);

CREATE TABLE arena.feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES arena.sessions(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES users.users(id),
    expert_id UUID NOT NULL REFERENCES users.users(id),
    comment TEXT NOT NULL CHECK (char_length(comment) BETWEEN 1 AND 10000),
    ratings JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_arena_feedback_session_id ON arena.feedback(session_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP SCHEMA IF EXISTS arena CASCADE;
DROP TABLE IF EXISTS training.session_messages;

-- +goose StatementEnd
