-- +goose Up
-- +goose StatementBegin

CREATE TABLE learning.media (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    media_type TEXT NOT NULL CHECK (media_type = 'voice_message'),
    content_type TEXT NOT NULL,
    payload BYTEA NOT NULL,
    duration_seconds INTEGER NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_owner_created ON learning.media(owner_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS learning.idx_media_owner_created;
DROP TABLE IF EXISTS learning.media;

-- +goose StatementEnd
