-- +goose Up
-- +goose StatementBegin

CREATE TABLE users.refresh_tokens (
      id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id     UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
      token_hash  CHAR(64) NOT NULL UNIQUE,
      active_role users.product_role NOT NULL DEFAULT 'learner',

      expires_at  TIMESTAMPTZ NOT NULL,
      revoked_at  TIMESTAMPTZ NULL,

      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      FOREIGN KEY (user_id, active_role)
            REFERENCES users.user_roles(user_id, role) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_id ON users.refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON users.refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_active_user ON users.refresh_tokens(user_id)
    WHERE revoked_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS users.idx_refresh_tokens_active_user;
DROP INDEX IF EXISTS users.idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS users.idx_refresh_tokens_user_id;
DROP TABLE IF EXISTS users.refresh_tokens;

-- +goose StatementEnd
