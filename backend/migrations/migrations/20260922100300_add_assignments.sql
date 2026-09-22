-- +goose Up
-- +goose StatementBegin

CREATE TABLE learning.assignments (
    id UUID PRIMARY KEY,
    curator_id UUID NOT NULL REFERENCES users.users(id),
    learner_id UUID NOT NULL REFERENCES users.users(id),
    scenario_id UUID NOT NULL REFERENCES scenarios.scenarios(id),
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','completed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX ux_assignments_pending ON learning.assignments(curator_id, learner_id, scenario_id) WHERE status = 'pending';
CREATE INDEX idx_assignments_curator_created ON learning.assignments(curator_id, created_at DESC);
CREATE INDEX idx_assignments_learner_created ON learning.assignments(learner_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS learning.idx_assignments_learner_created;
DROP INDEX IF EXISTS learning.idx_assignments_curator_created;
DROP INDEX IF EXISTS learning.ux_assignments_pending;
DROP TABLE IF EXISTS learning.assignments;

-- +goose StatementEnd
