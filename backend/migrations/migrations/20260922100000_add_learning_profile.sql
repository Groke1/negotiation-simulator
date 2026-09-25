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
 ('a0000000-0000-4000-8000-000000000001','STRATEGY_BOUNDARIES','Стратегия, цели и границы','Понимает свои цели и приоритеты, удерживает ограничения, не принимает заведомо плохие условия'),
 ('a0000000-0000-4000-8000-000000000002','DIAGNOSIS_LISTENING','Диагностика и слушание','Выясняет интересы, мотивы, ограничения, приоритеты; слышит ответы и использует полученную информацию'),
 ('a0000000-0000-4000-8000-000000000003','VALUE_CREATION','Создание ценности','Ищет дополнительные переменные, варианты и взаимовыгодные решения вместо торга только по одной величине'),
 ('a0000000-0000-4000-8000-000000000004','BARGAINING_EXCHANGE','Торг и обмен','Делает предложения и контрпредложения, управляет уступками, получает встречную ценность'),
 ('a0000000-0000-4000-8000-000000000005','ARGUMENTATION_INFLUENCE','Аргументация и влияние','Обосновывает позицию, использует критерии, работает с возражениями и framing'),
 ('a0000000-0000-4000-8000-000000000006','RELATIONSHIP_REGULATION','Отношения и эмоциональная регуляция','Сохраняет рабочие отношения, признаёт позицию второй стороны, снижает ненужную эскалацию'),
 ('a0000000-0000-4000-8000-000000000007','PROCESS_ADAPTABILITY','Управление процессом и адаптивность','Управляет направлением разговора, меняет тактику при новой информации, работает с тупиками'),
 ('a0000000-0000-4000-8000-000000000008','CLOSING_COMMITMENTS','Завершение и обязательства','Фиксирует договорённости, условия, ответственность и следующие шаги; корректно завершает без сделки')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE learning.semantic_actions (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO learning.semantic_actions (code, name, description) VALUES
 ('PROBE_INTEREST','Выяснить интерес','Понять, что на самом деле важно второй стороне'),
 ('PROBE_CONSTRAINT','Уточнить ограничение','Понять границы, ограничения, обязательные условия'),
 ('CLARIFY','Уточнить информацию','Получить недостающие факты или прояснить сказанное'),
 ('STATE_POSITION','Обозначить позицию','Сообщить своё требование, условие или предпочтение'),
 ('HOLD_BOUNDARY','Удержать границу','Показать, что дальше определённого условия пользователь не готов идти'),
 ('JUSTIFY','Обосновать позицию','Объяснить причины своей позиции или предложения'),
 ('MAKE_PROPOSAL','Сделать предложение','Предложить конкретный вариант решения'),
 ('MAKE_CONCESSION','Сделать уступку','Изменить свою позицию в пользу второй стороны'),
 ('CONDITIONAL_CONCESSION','Предложить обмен','«Если вы X, то мы Y» — уступка за встречную ценность'),
 ('ACKNOWLEDGE','Признать позицию собеседника','Показать, что позиция/интерес другой стороны услышаны'),
 ('CONFIRM_AGREEMENT','Зафиксировать договорённость','Подтвердить достигнутые условия'),
 ('END_NEGOTIATION','Завершить переговоры','Закончить переговоры с соглашением или без него'),
 ('OTHER','Другое','В реплике нет признаков ни одного действия из словаря или действие не относится к переговорной метке')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE learning.user_profile (
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    cluster_id UUID NOT NULL REFERENCES learning.competency_clusters(id),
    score DOUBLE PRECISION NOT NULL DEFAULT 50 CHECK (score >= 0 AND score <= 100),
    evidence_mass DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (evidence_mass >= 0),
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (confidence >= 0 AND confidence <= 100),
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
DROP TABLE IF EXISTS learning.semantic_actions;
DROP TABLE IF EXISTS learning.competency_clusters;
DROP SCHEMA IF EXISTS learning;

-- +goose StatementEnd
