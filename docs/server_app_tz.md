# ТЗ — сервер приложений (детализация к backend_tz.md)

> Дополняет backend_tz.md: там — схема БД, модули, список эндпоинтов. Здесь — контракты запросов/ответов, бизнес-логика по шагам, машина состояний сессии, обработка ошибок, нефункциональные требования.

---

## 1. Аутентификация и роли

- Bearer JWT в заголовке `Authorization`. Payload токена: `{ user_id, role }`, роль — одна из `learner` / `methodist` / `curator`.
- Middleware проверяет роль на группах маршрутов:

| Группа маршрутов | Разрешённые роли |
|---|---|
| `POST/PUT /api/scenarios*` (кроме чтения) | `methodist` |
| `GET /api/scenarios*` | любая роль |
| `POST /api/sessions*` (для себя, `user_id` из токена) | `learner` |
| `GET /api/users/{id}/profile`, `/recommendation` | сам пользователь (`learner`) либо `curator`, если `users.curator_id` этого пользователя = текущий |
| `GET /api/users/{id}/sessions`, `/profile/history` | то же правило, что и для `/profile` строкой выше — уточнено при добавлении модуля «Куратор» (изначально ошибочно подразумевался только сам `learner`) |
| `GET /api/curator/*`, `POST /api/assignments` | только `curator`, только в рамках своей команды (`curator_id` проверяется на каждой строке ответа, не только на входе в маршрут) |

**Проверка «в рамках своей команды»** — не разовая проверка на входе, а фильтр на уровне каждого запроса: куратор, обратившийся к `GET /api/users/{other_id}/profile` для сотрудника не из своей команды, получает `403`, а не пустой/чужой результат. Это касается всех маршрутов, где `curator_id` определяет доступ.

- Эндпоинт `POST /api/auth/login` — вне рамок этого ТЗ, если аутентификация на хакатоне упрощена (например, единый демо-пользователь на роль) — тогда JWT можно выдавать статично на старте демо; финальный вид логина — отдельная задача, не блокирует остальной API.

---

## 2. Машина состояний сессии

```
        POST /sessions
              │
              ▼
        in_progress ──────► completed   (POST /sessions/{id}/complete)
              │
              └──────────► abandoned    (таймаут бездействия ИЛИ явный вызов, опционально для MVP)
```

Только `in_progress`-сессия принимает `POST /sessions/{id}/events`. Попытка записать событие в `completed`/`abandoned` сессию → `409 Conflict`.

---

## 3. Контракты по эндпоинтам

### 3.1. Конструктор

**`POST /api/scenarios/generate`**

```json
// Запрос
{
  "prompt": "Создай тренировку для менеджера по продажам. Клиент хочет продлить контракт, требует скидку 15%..."
}
```
```json
// Ответ 200
{
  "scenario_draft": { /* полная структура Scenario из architecture_v3.md, раздел 6 */ },
  "ai_generation_id": "uuid",
  "provider": "cloud"
}
```
Бизнес-логика: вызвать `AIProvider.GenerateScenario(ctx, prompt)` → записать в `ai_generations` (raw_response) → **не сохранять** в `scenarios` автоматически — черновик возвращается клиенту для правки, сохраняется отдельным `POST /api/scenarios` уже отредактированным.

**`POST /api/scenarios/{id}/validate`**

```json
// Ответ 200
{
  "valid": false,
  "structure": { "has_goal": true, "has_conflict": true, "has_roles": true, "has_completion_conditions": true },
  "graph": { "dead_ends": ["n7"], "unreachable_nodes": [], "cycles": [] },
  "coverage": { "under_covered_indicators": [ { "indicator": "Запрашивает встречную ценность", "contexts_found": 1, "required": 3 } ] }
}
```
Код `200` даже при `valid: false` — это не ошибка запроса, а результат проверки. `400` — только если сам JSON сценария не парсится.

**`POST /api/scenarios/{id}/publish`**

- Отказ (`422`), если последняя `validate` для этой версии не проходилась или вернула `valid: false` — публикация невалидного сценария запрещена на уровне сервера, не только UI.
- При успехе — транзакционно: снимок в `scenario_versions`, `scenarios.status = 'published'`.

**`GET /api/scenarios/{id}/versions`**

```json
// Ответ 200
{
  "versions": [
    { "version": "1.1", "published_at": "2026-09-16T09:00:00Z" },
    { "version": "1.0", "published_at": "2026-09-15T18:00:00Z" }
  ]
}
```
```
// Просмотр конкретного снимка — тот же путь с версией, не отдельный эндпоинт:
GET /api/scenarios/{id}/versions/{version}
```
```json
// Ответ 200
{ "version": "1.0", "snapshot": { /* полная структура Scenario на момент публикации этой версии */ }, "published_at": "..." }
```
Простая проекция по `scenario_versions`, сортировка `published_at DESC`. Используется экраном 9 «История версий» в frontend_constructor_tz.md.

### 3.2. Обучение

**`POST /api/sessions`**

```json
// Запрос
{ "scenario_id": "uuid", "channel": "chat", "mode": "training", "preview": false }
```
- `preview` — необязательное поле, по умолчанию `false`. `true` разрешено только для роли `methodist` (используется экраном 7 «Предпросмотр» в frontend_constructor_tz.md) — попытка передать `preview: true` от роли `learner` → `403`.
- Сессии с `preview: true` **не участвуют** в шаге 4 «Scoring» при завершении (раздел ниже) — не пишутся в `profile_indicator_levels`/`profile_indicator_history`, не порождают запись в `recommendation_log`. Событие всё равно записывается в `events` (для работы самого Event Inspector на экране предпросмотра), но помечено принадлежностью к превью-сессии — фильтруется на уровне `sessions.is_preview`, не отдельной таблицей.
```json
// Ответ 201
{
  "session_id": "uuid",
  "scenario_version": "2.0",
  "current_node": {
    "node_id": "n1",
    "speaker": "partner",
    "text": "Добрый день. Готовы продлить договор, но хотели бы снизить стоимость на 15%.",
    "available_actions": [
      { "action_id": "a1", "label": "Уточнить причины", "semantic_action": "ASK_QUESTION" },
      { "action_id": "a2", "label": "Предложить обмен", "semantic_action": "MAKE_PROPOSAL" }
    ]
  }
}
```

**`POST /api/sessions/{id}/events`**

```json
// Запрос
{ "action_id": "a2", "client_event_id": "uuid" }
```
- `client_event_id` — обязателен, используется для идемпотентности: уникальный индекс `(session_id, client_event_id)`. Критично для канала Voice — если STT/сеть дублирует отправку, повторный вызов с тем же `client_event_id` должен возвращать уже сохранённый результат, а не создавать второе событие.
```json
// Ответ 200
{
  "event_id": "uuid",
  "next_node": { /* тот же формат, что в ответе POST /sessions выше */ },
  "session_status": "in_progress"
}
```
- Если `next_node` отсутствует (достигнуто завершение по `completion_conditions`) — `next_node: null`, `session_status: "ready_to_complete"` — клиент должен вызвать `/complete`.
- `422`, если `action_id` не входит в `available_actions` текущего узла (защита от рассинхронизации клиента).

**`POST /api/sessions/{id}/complete`**

Бизнес-логика по шагам (транзакция):

0. Если `sessions.is_preview = true` — выполнить только шаги 1–2 (прочитать события, вычислить `outcome`), пропустить шаги 3–6 (Scoring, обновление профиля, история, рекомендация) целиком. Ответ клиенту — тот же формат, но без `profile_updated` (пустой массив).
1. Прочитать все `events` сессии по `sequence_no`.
2. Вычислить `outcome` — сопоставить финальное состояние параметров сделки с `completion_conditions` сценария.
3. **Scoring**: для каждого события сопоставить `verb` + `context` с `observation_rules` версии сценария (снимок из `scenario_versions`, не текущий черновик — правило неизменности результата), получить список подтверждённых индикаторов.
4. Обновить `profile_indicator_levels` (upsert по `(user_id, indicator_id)`, источник `'training'` или `'arena'` из `sessions.mode`).
5. Вставить строки в `profile_indicator_history` (по одной на каждое изменение уровня) — источник данных для «до/после».
6. Вызвать Recommendation Engine, записать результат в `recommendation_log`.
7. `sessions.status = 'completed'`, `completed_at = now()`.

```json
// Ответ 200
{
  "outcome": { "contract_extended": true, "price_change": -0.10 },
  "profile_updated": [
    { "indicator": "Запрашивает встречную ценность", "previous_level": 1, "new_level": 2 }
  ]
}
```

**`GET /api/sessions/{id}/analysis`**

```json
{
  "result": { "contract_extended": true, "price_change": -0.10, "goals_met": ["term"], "goals_missed": ["price"] },
  "process": [
    { "step": 1, "actor": "partner", "summary": "Клиент назвал ограничения" },
    { "step": 2, "actor": "learner", "verb": "MAKE_CONCESSION", "summary": "Предложили скидку 10%", "flag": "без встречной ценности" }
  ],
  "key_moment": { "step": 2, "note": "После первой уступки была возможность обменять её на срок контракта — не использована" },
  "competency": [
    { "cluster": "Торг и обмен", "indicator": "Запрашивает встречную ценность", "level": 2, "note": "уступали до проверки возможности обмена" }
  ]
}
```

Формулировки в поле `note`/`summary` формируются на сервере по шаблонам (rules-based текст из зафиксированных наблюдений), не LLM — соответствует принципу «ИИ не в рантайме обучающегося».

**`GET /api/users/{id}/profile`**, **`GET /api/users/{id}/recommendation`** — простые проекции из `profile_indicator_levels` и `recommendation_log`, без отдельной сложной логики сверх Recommendation Engine, описанного в backend_tz.md раздел 5.

**`GET /api/users/{id}/sessions`**

```
// Query-параметры
?mode=training|arena   (опционально, без параметра — все режимы)
?status=completed       (по умолчанию — только completed, для истории незавершённые не нужны)
```
```json
// Ответ 200
{
  "sessions": [
    { "session_id": "uuid", "scenario_title": "Продление контракта — ценовой конфликт", "mode": "arena", "channel": "chat", "completed_at": "2026-09-17T10:00:00Z" }
  ]
}
```
Простая проекция по таблице `sessions` (JOIN на `scenarios` для `title`), сортировка по `completed_at DESC`. Используется экраном 8 «История тренировок» в frontend_training_tz.md и arena_frontend_tz.md — клик по элементу списка ведёт на уже существующий `GET /api/sessions/{id}/analysis`, отдельного эндпоинта не требуется.

### 3.3. Арена

**`POST /api/sessions/{id}/preparation`**

```json
// Запрос
{ "goal": "Сохранить контракт на 12 месяцев", "min_acceptable": "6 месяцев", "priorities": ["срок","цена","SLA"], "possible_exchanges": ["цена ↔ объём"] }
```
- Разрешён только для сессии со статусом `in_progress`, `mode='arena'`, и только **до** первого записанного события — `409`, если события уже есть.

### 3.4. Куратор

**`GET /api/curator/team`**

```json
// Ответ 200
{
  "team": [
    { "user_id": "uuid", "name": "Иван Петров", "last_session_at": "2026-09-15T10:00:00Z", "weakest_cluster": "Торг и обмен" }
  ]
}
```
`weakest_cluster` вычисляется на сервере как кластер с наименьшим средним уровнем по `profile_indicator_levels` среди `is_mvp = true` индикаторов этого пользователя — не отдельное поле в БД, а проекция при чтении.

**`GET /api/users/{id}/profile/history`**

```
// Query-параметры
?from=2026-08-01&to=2026-09-17
// либо
?session_id_from=uuid&session_id_to=uuid
```
```json
// Ответ 200
{
  "clusters": [
    {
      "cluster": "Диагностика",
      "before": { "indicator": "Задаёт вопросы о потребностях/интересах", "level": 1 },
      "after":  { "indicator": "Задаёт вопросы о потребностях/интересах", "level": 3 }
    }
  ]
}
```
Источник — `profile_indicator_history`, ближайшая запись к каждой границе диапазона по `recorded_at`. Формулировки словами («было эпизодически, стало стабильно») формируются на клиенте по уровню 0–4, не на сервере — сервер отдаёт только числа и названия индикаторов.

**`POST /api/assignments`**

```json
// Запрос
{ "learner_id": "uuid", "scenario_id": "uuid" }
```
- `403`, если `learner_id` не находится в команде текущего куратора (`users.curator_id != request.user_id`).
- `404`, если `scenario_id` не найден или не `status='published'` — назначить черновик нельзя.
```json
// Ответ 201
{ "assignment_id": "uuid", "status": "pending" }
```
Влияет на выдачу главного экрана «Обучения» — эндпоинт `GET /api/users/{id}/recommendation` должен при наличии `pending`-назначения возвращать его вместо (или наряду с) обычной rules-based рекомендацией, с пометкой источника (`"source": "curator_assignment"` vs `"source": "recommendation_engine"`) — фронтенд использует эту пометку для отображения «назначено куратором» (frontend_curator_tz.md, экран 5).

**`GET /api/curator/team/analytics`**

```json
// Ответ 200
{
  "clusters": [
    {
      "cluster": "Торг и обмен",
      "distribution": { "level_0_1": 2, "level_2_3": 5, "level_4": 1 }
    }
  ],
  "most_common_weak_cluster": "Торг и обмен"
}
```
Агрегат по `profile_indicator_levels` всех `learner`, где `curator_id = request.user_id` — `GROUP BY cluster_id`, подсчёт по корзинам уровня. Простой SQL-агрегат, без отдельного сервиса.

---

## 4. Формат ошибок (единый по всем эндпоинтам)

```json
{
  "error": {
    "code": "ACTION_NOT_AVAILABLE",
    "message": "Действие a5 недоступно в текущем узле n3",
    "details": { "available_action_ids": ["a1", "a2"] }
  }
}
```

| HTTP-код | Когда |
|---|---|
| 400 | Невалидный JSON запроса |
| 401 | Отсутствует/невалиден JWT |
| 403 | Роль не соответствует маршруту |
| 404 | Сущность не найдена |
| 409 | Конфликт состояния (событие в завершённую сессию, повторная Подготовка после старта) |
| 422 | Семантически невалидный запрос (`action_id` не из текущего узла, публикация невалидного сценария) |
| 500 | Внутренняя ошибка — логируется с `request_id`, наружу — только код без деталей |

---

## 5. Нефункциональные требования

- **Идемпотентность** — `client_event_id` обязателен на `/events` (раздел 3.2). Без этого Voice-канал (потенциальные повторные срабатывания STT) может задвоить события и исказить Scoring.
- **Rate limiting на `/api/scenarios/generate`** — простой счётчик в Postgres (таблица `ai_generations` уже логирует каждый вызов; лимит — N генераций на пользователя в сутки, проверка перед вызовом провайдера) — прямое следствие free_tier_plan.md: контроль бюджета на уровне сервера, не только по факту счёта от провайдера.
- **`GET /healthz`** — без аутентификации, проверяет соединение с БД, нужен для Railway/Render (implementation_roadmap.md, раздел про публичный доступ).
- **Конфигурация через переменные окружения**: `DATABASE_URL`, `AI_PROVIDER` (`mock`/`cloud`/`ollama`), `AI_API_KEY`, `JWT_SECRET`, `CORS_ORIGINS`, `PORT` — ничего не хардкодится, соответствует free_tier_plan.md.
- **Миграции — отдельная команда** (`migrate up`), не выполняются автоматически при старте сервера — безопаснее при повторных деплоях на демо.
- **CORS** — разрешённые origin'ы только для доменов веб-клиента и Author Console (`CORS_ORIGINS` как список, не `*`).
- **Логирование** — каждый запрос: `request_id`, `user_id` (если есть), маршрут, код ответа, время выполнения. Структурировано (zerolog), не текстом.
