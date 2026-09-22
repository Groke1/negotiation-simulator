-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS learning;

CREATE TABLE learning.competency_clusters (
    id UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO learning.competency_clusters (id, code, name, description) VALUES
 ('a0000000-0000-4000-8000-000000000001','goals','Цели и границы','Постановка целей, удержание рамки и критериев результата'),
 ('a0000000-0000-4000-8000-000000000002','diag','Диагностика','Выявление интересов, ограничений и контекста партнера'),
 ('a0000000-0000-4000-8000-000000000003','value','Создание ценности','Поиск взаимной выгоды и расширение поля вариантов'),
 ('a0000000-0000-4000-8000-000000000004','trade','Торг и обмен','Предложения, уступки и обмен ценностями'),
 ('a0000000-0000-4000-8000-000000000005','adapt','Адаптивность','Гибкость поведения и реакция на изменения'),
 ('a0000000-0000-4000-8000-000000000006','close','Завершение','Фиксация договоренностей и итоговых шагов')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE learning.user_profile (
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    cluster_id UUID NOT NULL REFERENCES learning.competency_clusters(id),
    score SMALLINT NOT NULL DEFAULT 50 CHECK (score >= 0 AND score <= 100),
    updated_at TIMESTAMPTZ NULL,
    PRIMARY KEY (user_id, cluster_id)
);

CREATE TABLE learning.profile_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    source TEXT NOT NULL CHECK (source IN ('training','arena_feedback')),
    source_id UUID NOT NULL,
    changes JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profile_history_user_created ON learning.profile_history(user_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS learning.idx_profile_history_user_created;
DROP TABLE IF EXISTS learning.profile_history;
DROP TABLE IF EXISTS learning.user_profile;
DROP TABLE IF EXISTS learning.competency_clusters;
DROP SCHEMA IF EXISTS learning;

-- +goose StatementEnd
