# ТЗ для бэкендера — PostgreSQL + Go

> Основано на architecture_v3.md. Структура ровно по трём логическим модулям: **Конструктор** (создание Scenario), **Обучение** (основной цикл — 90% продукта), **Арена** (опциональный редкий элемент, схема готова с запасом под Живую Арену, но эндпоинты для неё — не в хакатонном объёме).

---

## 0. Общие решения

| Вопрос | Решение | Почему |
|---|---|---|
| Язык | Go 1.22+ | Задано |
| БД | PostgreSQL 15+ | Задано |
| Генерация UUID | На стороне приложения (`github.com/google/uuid`), не `pgcrypto` | Меньше зависимости от версии/расширений СУБД, проще переносить между Railway/Render/локалью |
| Формат id | UUID v4 везде, включая внешние ключи | Единообразно с `session_id`/`scenario_id`, которые уже фигурируют как `uuid` в JSON-схемах architecture_v3.md |
| Хранение структур Scenario/Session | `jsonb`-колонки поверх реляционного каркаса, не EAV и не полностью документная БД | Structured-часть (id, статус, владелец, даты) — для быстрых выборок и индексов; вложенная структура (roles, goals, nodes и т.д.) — не нуждается в отдельных таблицах на MVP, JSON Schema из architecture_v3.md и так уже описывает её форму |

---

## 1. Рекомендуемые библиотеки

| Назначение | Библиотека | Комментарий |
|---|---|---|
| HTTP-роутер | [`go-chi/chi`](https://github.com/go-chi/chi) | Тонкий, идиоматичный, совместим с `net/http` напрямую — не тянет вес отдельного фреймворка. Если команде важнее скорость написания хендлеров, чем идиоматичность — альтернатива [`gin-gonic/gin`](https://github.com/gin-gonic/gin), тоже приемлемо |
| Драйвер PostgreSQL | [`jackc/pgx/v5`](https://github.com/jackc/pgx) (`pgxpool` для пула соединений) | Быстрее `database/sql` + `lib/pq`, нативная поддержка `jsonb` без ручного маршалинга в отдельный тип |
| Типобезопасные запросы к БД | [`sqlc`](https://sqlc.dev/) | Генерирует Go-код из `.sql` — запросы остаются читаемым SQL (важно при активной работе с `jsonb`), а не спрятаны за ORM-магией. Если команда торопится и предпочитает ORM — альтернатива [`GORM`](https://gorm.io/), но тогда `jsonb`-поля потребуют ручных `Scan`/`Value` |
| Миграции схемы | [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) | Стандарт де-факто, CLI + библиотека, простой формат `.up.sql`/`.down.sql` |
| Валидация входных payload | [`go-playground/validator/v10`](https://github.com/go-playground/validator) | Теги валидации прямо на структурах запроса |
| Валидация структуры Scenario-JSON | [`santhosh-tekuri/jsonschema/v5`](https://github.com/santhosh-tekuri/jsonschema) | Нужен отдельно от validator — это не валидация формы запроса, а обязательный шаг «AI-проверка сценария» из architecture_v3.md (структура/логика графа) |
| Конфигурация из переменных окружения | [`caarlos0/env/v10`](https://github.com/caarlos0/env) | Совпадает с free_tier_plan.md — ключи ИИ-провайдера и параметры БД через env, ничего не хардкодится |
| Логирование | [`rs/zerolog`](https://github.com/rs/zerolog) | Структурированные логи, минимальные накладные расходы |
| JWT-аутентификация | [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt) | Минимально необходимо различать обучающегося/методиста/куратора между запросами |
| CORS | [`go-chi/cors`](https://github.com/go-chi/cors) | Веб-клиент и Author Console — разные источники (см. implementation_roadmap.md, раздел про Vercel/Netlify) |
| Тесты | стандартный `testing` + [`stretchr/testify`](https://github.com/stretchr/testify) | `assert`/`require`, `httptest` для эндпоинтов |
| HTTP-клиент к AI-провайдеру | стандартный `net/http` + официальный SDK провайдера, если есть (например `anthropic-sdk-go`) | Не нужен отдельный HTTP-фреймворк ради одного внешнего вызова |

---

## 2. Структура Go-проекта — по трём модулям

```
/cmd/server/main.go

/internal/
  /platform          — конфиг, логгер, пул БД, middleware (auth, CORS, request-id)
  /shared            — общий для всех трёх модулей код:
                          • словарь семантических действий (единственный источник правды,
                            не дублировать между пакетами)
                          • интерфейс AIProvider (Mock / Cloud / Ollama)
                          • Event Log — запись/чтение событий (используется и Обучением, и Ареной)
  /constructor       — модуль «Конструктор»
  /training          — модуль «Обучение»
  /arena             — модуль «Арена» (использует /shared/eventlog и /training напрямую,
                         не дублирует логику сессий)

/db/migrations       — golang-migrate .sql файлы
/db/queries          — sqlc .sql запросы, разложены по подпапкам constructor/ training/ arena/
```

Важно: `/arena` **не копирует** логику `/training` — модели `Session`/`Event` в базе одни и те же для обоих режимов (различаются только полем `mode`), поэтому Арена в Go-коде — тонкая надстройка над общими сервисами `/shared` и `/training`, а не параллельная реализация. Это прямое следствие того, что в architecture_v3.md Арена — не отдельная система, а вариант Практики.

---

## 3. Схема PostgreSQL

### 3.1. Общие таблицы (`/shared`, используются всеми тремя модулями)

```sql
CREATE TABLE users (
    id          uuid PRIMARY KEY,
    email       text UNIQUE NOT NULL,
    name        text NOT NULL,
    role        text NOT NULL CHECK (role IN ('learner', 'methodist', 'curator')),
    curator_id  uuid REFERENCES users(id),   -- заполняется только для role='learner'; MVP: один куратор на сотрудника
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_curator ON users(curator_id);

CREATE TABLE competency_clusters (
    id   smallserial PRIMARY KEY,
    name text UNIQUE NOT NULL   -- 6 кластеров: Цели и границы, Диагностика, ...
);

CREATE TABLE indicators (
    id                    serial PRIMARY KEY,
    cluster_id            smallint NOT NULL REFERENCES competency_clusters(id),
    name                  text NOT NULL,
    operational_definition text NOT NULL,
    automatable           text NOT NULL CHECK (automatable IN ('high', 'medium', 'low')),
    is_mvp                boolean NOT NULL DEFAULT false
);
```

`indicators` — справочник, заполняется один раз миграцией данных (seed) из матрицы 42 индикаторов; поле `is_mvp` отделяет детерминированно вычислимое MVP-подмножество (architecture_v3.md, раздел 5) от остального.

### 3.2. Модуль «Конструктор»

```sql
CREATE TABLE scenarios (
    id                    uuid PRIMARY KEY,
    version               text NOT NULL,
    title                 text NOT NULL,
    status                text NOT NULL CHECK (status IN ('draft', 'published')) DEFAULT 'draft',
    mode_support          text[] NOT NULL DEFAULT '{}',   -- ['training','arena','live']
    context               jsonb NOT NULL,
    roles                 jsonb NOT NULL,
    goals                 jsonb NOT NULL,
    interests             jsonb NOT NULL,
    constraints           jsonb NOT NULL,
    parameters            jsonb NOT NULL,
    private_information   jsonb NOT NULL DEFAULT '{}',
    difficulty            text NOT NULL,
    competencies          text[] NOT NULL DEFAULT '{}',
    observation_rules     jsonb NOT NULL DEFAULT '[]',
    completion_conditions jsonb NOT NULL DEFAULT '[]',
    channels              text[] NOT NULL DEFAULT '{}',   -- ['chat','voice']
    nodes                 jsonb NOT NULL,                 -- граф для направляемой тренировки/Арены
    created_by            uuid NOT NULL REFERENCES users(id),
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- Неизменяемые снимки опубликованных версий — правило "изменение сценария
-- не должно менять результаты уже пройденных тренировок" (architecture_v2/3)
CREATE TABLE scenario_versions (
    id            uuid PRIMARY KEY,
    scenario_id   uuid NOT NULL REFERENCES scenarios(id),
    version       text NOT NULL,
    snapshot      jsonb NOT NULL,      -- полная копия scenarios.* на момент публикации
    published_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (scenario_id, version)
);

-- Лог обращений к AI Scenario Builder — нужен и для отладки, и для контроля
-- бюджета CloudAIProvider (free_tier_plan.md)
CREATE TABLE ai_generations (
    id                uuid PRIMARY KEY,
    scenario_id       uuid REFERENCES scenarios(id),   -- NULL, пока черновик не сохранён
    author_id         uuid NOT NULL REFERENCES users(id),
    provider          text NOT NULL CHECK (provider IN ('mock', 'cloud', 'ollama')),
    input_prompt      text NOT NULL,
    raw_response      jsonb NOT NULL,
    validation_result jsonb,                            -- результат шага "AI-проверка сценария"
    created_at        timestamptz NOT NULL DEFAULT now()
);
```

### 3.3. Модуль «Обучение» (используется и «Ареной» без изменений схемы)

```sql
CREATE TABLE sessions (
    id                uuid PRIMARY KEY,
    scenario_id       uuid NOT NULL REFERENCES scenarios(id),
    scenario_version  text NOT NULL,
    user_id           uuid NOT NULL REFERENCES users(id),   -- обучающийся
    mode              text NOT NULL CHECK (mode IN ('training', 'arena', 'live')),
    channel           text NOT NULL CHECK (channel IN ('chat', 'voice', 'meeting')),
    status            text NOT NULL CHECK (status IN ('in_progress', 'completed', 'abandoned')) DEFAULT 'in_progress',
    participants      jsonb NOT NULL,       -- [{role, actor_type, actor_id}]
    judge_id          uuid REFERENCES users(id),   -- только для mode='live', иначе NULL
    preparation        jsonb,                -- рабочий лист перед Ареной; NULL для mode='training'
    is_preview        boolean NOT NULL DEFAULT false,  -- сессии методиста на экране «Предпросмотр» — не участвуют в Scoring
    outcome           jsonb,
    started_at        timestamptz NOT NULL DEFAULT now(),
    completed_at      timestamptz
);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_scenario ON sessions(scenario_id);

-- Event Log — xAPI-подобный формат из architecture_v3.md, один на все режимы и каналы
CREATE TABLE events (
    id           uuid PRIMARY KEY,
    session_id   uuid NOT NULL REFERENCES sessions(id),
    sequence_no  int NOT NULL,             -- порядок внутри сессии
    actor        text NOT NULL,            -- 'learner' | 'virtual_partner' | 'partner' | uuid реального участника
    verb         text NOT NULL,            -- семантическое действие: ASK_QUESTION, MAKE_CONCESSION, ...
    object       jsonb NOT NULL,           -- {node_id, action_id}
    context      jsonb NOT NULL,           -- {channel, state_before, state_after}
    result       jsonb,                    -- {indicator_tags: [...]}
    occurred_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (session_id, sequence_no)
);
CREATE INDEX idx_events_session ON events(session_id);

-- Текущий уровень по индикатору — агрегат, обновляется после каждой завершённой сессии
CREATE TABLE profile_indicator_levels (
    id            uuid PRIMARY KEY,
    user_id       uuid NOT NULL REFERENCES users(id),
    indicator_id  int NOT NULL REFERENCES indicators(id),
    level         smallint NOT NULL CHECK (level BETWEEN 0 AND 4),
    source        text NOT NULL,          -- 'training' | 'arena' | 'live_judge' | 'live_perception' | ...
    weight        numeric NOT NULL DEFAULT 1.0,
    confirmed_at  timestamptz,
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, indicator_id)
);

-- История изменений уровня — источник для сравнения "до/после" и для календаря пилота
CREATE TABLE profile_indicator_history (
    id            uuid PRIMARY KEY,
    user_id       uuid NOT NULL REFERENCES users(id),
    indicator_id  int NOT NULL REFERENCES indicators(id),
    level         smallint NOT NULL CHECK (level BETWEEN 0 AND 4),
    source        text NOT NULL,
    session_id    uuid NOT NULL REFERENCES sessions(id),
    recorded_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_profile_history_user ON profile_indicator_history(user_id);

-- Что и почему было рекомендовано — для отладки Recommendation Engine и для куратора
CREATE TABLE recommendation_log (
    id                     uuid PRIMARY KEY,
    user_id                uuid NOT NULL REFERENCES users(id),
    recommended_scenario_id uuid NOT NULL REFERENCES scenarios(id),
    reason                 jsonb NOT NULL,   -- {weak_indicators: [...]}
    created_at             timestamptz NOT NULL DEFAULT now()
);

-- Явное назначение тренировки куратором — отдельно от автоматической рекомендации,
-- чтобы на экране 1 "Обучения" можно было различить "система посоветовала" и "куратор поручил"
CREATE TABLE assignments (
    id            uuid PRIMARY KEY,
    curator_id    uuid NOT NULL REFERENCES users(id),
    learner_id    uuid NOT NULL REFERENCES users(id),
    scenario_id   uuid NOT NULL REFERENCES scenarios(id),
    status        text NOT NULL CHECK (status IN ('pending', 'completed')) DEFAULT 'pending',
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_assignments_learner ON assignments(learner_id) WHERE status = 'pending';
```

### 3.4. Модуль «Арена» — только то, чего нет в «Обучении»

Арена как режим уже полностью покрыта таблицами `sessions`/`events` выше (`mode = 'arena'`, поле `preparation` заполнено). Дополнительно нужны только таблицы под будущую Живую Арену — **создать схемой сейчас, эндпоинты не реализовывать в хакатоне** (architecture_v3.md прямо выносит это за рамки MVP, но откладывать миграцию на потом дороже, чем создать пустые таблицы сразу):

```sql
CREATE TABLE live_intentions (
    id                uuid PRIMARY KEY,
    session_id        uuid NOT NULL REFERENCES sessions(id),
    event_id          uuid NOT NULL REFERENCES events(id),
    actor_id          uuid NOT NULL REFERENCES users(id),
    stated_intention  text NOT NULL,
    created_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE live_perceptions (
    id             uuid PRIMARY KEY,
    session_id     uuid NOT NULL REFERENCES sessions(id),
    event_id       uuid NOT NULL REFERENCES events(id),
    actor_id       uuid NOT NULL REFERENCES users(id),
    perceived_as   text NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE judge_observations (
    id              uuid PRIMARY KEY,
    session_id      uuid NOT NULL REFERENCES sessions(id),
    event_id        uuid REFERENCES events(id),   -- NULL, если наблюдение не привязано к одному событию
    judge_id        uuid NOT NULL REFERENCES users(id),
    observation     text NOT NULL,
    indicator_tags  jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);
```

---

## 4. Эндпоинты — по модулям

### 4.1. Конструктор

| Метод | Путь | Назначение |
|---|---|---|
| POST | `/api/scenarios` | Создать черновик вручную |
| POST | `/api/scenarios/generate` | Сгенерировать черновик через `AIProvider` (тело: текстовое описание задачи) |
| GET | `/api/scenarios` | Библиотека тренировок (список, фильтр по `status`) |
| GET | `/api/scenarios/{id}` | Получить сценарий |
| PUT | `/api/scenarios/{id}` | Отредактировать черновик |
| POST | `/api/scenarios/{id}/validate` | Проверка структуры/логики графа **и** покрытия индикаторов (jsonschema + обход графа на тупики/циклы/недостижимые ветки + подсчёт «≥3 контекста на индикатор») — единый вызов, отдельного `/coverage`-эндпоинта нет (уточнено при сверке с server_app_tz.md/frontend_constructor_tz.md, экран 6) |
| POST | `/api/scenarios/{id}/publish` | Публикация — создаёт запись в `scenario_versions`, меняет `status` |
| GET | `/api/scenarios/{id}/versions` | Список опубликованных версий |
| GET | `/api/scenarios/{id}/versions/{version}` | Снимок конкретной версии (для frontend_constructor_tz.md, экран 9) |

### 4.2. Обучение

| Метод | Путь | Назначение |
|---|---|---|
| POST | `/api/sessions` | Начать сессию (`mode='training'`, тело: `scenario_id`, `channel`, опционально `preview` — только роль `methodist`, см. раздел 3.3) |
| GET | `/api/sessions/{id}/state` | Текущий узел и доступные действия (сервер — источник правды о состоянии, не клиент) |
| POST | `/api/sessions/{id}/events` | Записать совершённое действие (`verb` = семантическое действие) |
| POST | `/api/sessions/{id}/complete` | Завершить сессию → запускает Scoring → обновляет `profile_indicator_levels` (пропускается, если `is_preview = true`) |
| GET | `/api/sessions/{id}/analysis` | Разбор: Результат / Процесс / Компетенция |
| GET | `/api/users/{id}/profile` | Текущая карта компетенций |
| GET | `/api/users/{id}/recommendation` | Следующая рекомендованная тренировка |
| GET | `/api/users/{id}/sessions` | История сессий (фильтры `?mode=`, `?status=`) — для экрана «История тренировок» |

### 4.3. Арена

| Метод | Путь | Назначение |
|---|---|---|
| POST | `/api/sessions` | То же самое, `mode='arena'` — общий эндпоинт с Обучением, не дублируется |
| POST | `/api/sessions/{id}/preparation` | Сохранить рабочий лист (цель/границы/интересы) перед стартом |

Живая Арена — без эндпоинтов в хакатонном объёме; таблицы созданы миграцией заранее (раздел 3.4), API — Фаза 2+ (implementation_roadmap.md).

### 4.4. Куратор

| Метод | Путь | Назначение |
|---|---|---|
| GET | `/api/curator/team` | Список сотрудников (`users.curator_id = текущий пользователь`) с предвычисленным слабым кластером на строку — для frontend_curator_tz.md, экран 1 |
| GET | `/api/users/{id}/sessions` | Тот же эндпоинт, что в разделе 4.2 — доступ роли `curator` для сотрудников своей команды (проверка `curator_id` на сервере, не только `learner` для себя) |
| GET | `/api/users/{id}/profile/history` | Выборка из `profile_indicator_history` по диапазону дат или паре `session_id` — для экрана 4 «Сравнение до/после» |
| POST | `/api/assignments` | Назначить тренировку сотруднику (тело: `learner_id`, `scenario_id`) — доступно только роли `curator`, только для своей команды |
| GET | `/api/curator/team/analytics` | Агрегат по `profile_indicator_levels` всех сотрудников команды, группировка по `cluster_id` — для экрана 6 «Аналитика по команде» |

---

## 5. Scoring и Recommendation — где именно это код, а не ИИ

Оба сервиса живут в `/internal/training` и работают **только** с уже записанными `events`, без обращения к AI-провайдеру:

- **Scoring**: по завершении сессии читает все `events` данной `session_id`, сопоставляет `verb` + `context` с `observation_rules` соответствующего `scenario_versions.snapshot`, инкрементально обновляет `profile_indicator_levels` (правило весов по источнику — architecture_v3.md, раздел 10).
- **Recommendation**: читает `profile_indicator_levels` пользователя, находит индикаторы с наименьшим уровнем среди `is_mvp = true`, ищет в `scenarios` (status='published') сценарий, чей массив `competencies` покрывает нужный кластер — простое rules-based правило, не ML.

Это прямое продолжение free_tier_plan.md: единственная точка, где нужен платный AI — `/api/scenarios/generate`, всё остальное в бэкенде — обычный детерминированный код.

---

## 6. Что не входит в это ТЗ

- Реализация `AIProvider` для конкретного облачного провайдера (ключ, эндпоинт) — фиксируется отдельно при подключении `CloudAIProvider`, интерфейс уже описан в free_tier_plan.md.
- Live Arena API — только схема БД, без хендлеров (раздел 3.4).
- Синхронизация с офлайн-клиентом (WorkManager/мобильный сценарий) — относится к постхакатонной Фазе 2 из implementation_roadmap.md, не к вебовому MVP.
