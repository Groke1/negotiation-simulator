-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS users;

CREATE TYPE users.product_role AS ENUM (
    'learner',
    'methodist',
    'curator'
);

CREATE TABLE users.users (
     id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     organization_id UUID NULL REFERENCES organizations.organizations(id) ON DELETE SET NULL,

     username        VARCHAR(100) NOT NULL,
     email           VARCHAR(255) NOT NULL,
     password_hash   VARCHAR(255) NOT NULL,

     created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_users_username ON users.users (LOWER(username));
CREATE UNIQUE INDEX ux_users_email ON users.users (LOWER(email));
CREATE INDEX idx_users_organization_id ON users.users (organization_id);

CREATE TABLE users.user_roles (
      user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
      role users.product_role NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      PRIMARY KEY (user_id, role)
);

CREATE INDEX idx_user_roles_role ON users.user_roles (role);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS users.user_roles;
DROP INDEX IF EXISTS users.idx_users_organization_id;
DROP INDEX IF EXISTS users.ux_users_email;
DROP INDEX IF EXISTS users.ux_users_username;
DROP TABLE IF EXISTS users.users;
DROP TYPE IF EXISTS users.product_role;
DROP SCHEMA IF EXISTS users;

-- +goose StatementEnd
