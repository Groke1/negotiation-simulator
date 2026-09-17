-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS organizations;

CREATE TABLE organizations.organizations (
     id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name       VARCHAR(200) NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_organizations_name ON organizations.organizations (LOWER(name));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS organizations.organizations;
DROP SCHEMA IF EXISTS organizations;

-- +goose StatementEnd