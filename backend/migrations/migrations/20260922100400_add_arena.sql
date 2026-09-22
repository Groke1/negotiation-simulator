-- +goose Up
-- +goose StatementBegin

CREATE TABLE learning.arena_assignments (
    id UUID PRIMARY KEY,
    scenario_id UUID NOT NULL REFERENCES scenarios.scenarios(id),
    created_by UUID NOT NULL REFERENCES users.users(id),
    participants JSONB NOT NULL,
    channel TEXT NOT NULL DEFAULT 'chat' CHECK(channel IN ('chat','voice')),
    scheduled_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK(status IN ('scheduled','in_progress','completed','cancelled')),
    session_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_arena_assignments_created_by ON learning.arena_assignments(created_by, created_at DESC);
CREATE INDEX idx_arena_assignments_participants ON learning.arena_assignments USING gin(participants);

CREATE TABLE learning.arena_sessions (
    id UUID PRIMARY KEY,
    assignment_id UUID NOT NULL UNIQUE REFERENCES learning.arena_assignments(id),
    status TEXT NOT NULL DEFAULT 'waiting_for_participants' CHECK(status IN ('waiting_for_participants','in_progress','completed','cancelled')),
    joined JSONB NOT NULL DEFAULT '[]'::jsonb,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL
);

CREATE TABLE learning.arena_messages (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES learning.arena_sessions(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users.users(id),
    text TEXT NOT NULL,
    semantic_action_code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_arena_messages_session_created ON learning.arena_messages(session_id, created_at);

CREATE TABLE learning.arena_feedback (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES learning.arena_sessions(id),
    participant_id UUID NOT NULL REFERENCES users.users(id),
    curator_id UUID NOT NULL REFERENCES users.users(id),
    comment TEXT NOT NULL,
    ratings JSONB NOT NULL DEFAULT '[]'::jsonb,
    profile_change JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, participant_id)
);

CREATE INDEX idx_arena_feedback_session ON learning.arena_feedback(session_id);

ALTER TABLE learning.arena_assignments
    ADD CONSTRAINT fk_arena_assignments_session
    FOREIGN KEY (session_id) REFERENCES learning.arena_sessions(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE learning.arena_assignments DROP CONSTRAINT IF EXISTS fk_arena_assignments_session;
DROP INDEX IF EXISTS learning.idx_arena_feedback_session;
DROP TABLE IF EXISTS learning.arena_feedback;
DROP INDEX IF EXISTS learning.idx_arena_messages_session_created;
DROP TABLE IF EXISTS learning.arena_messages;
DROP TABLE IF EXISTS learning.arena_sessions;
DROP INDEX IF EXISTS learning.idx_arena_assignments_participants;
DROP INDEX IF EXISTS learning.idx_arena_assignments_created_by;
DROP TABLE IF EXISTS learning.arena_assignments;

-- +goose StatementEnd
