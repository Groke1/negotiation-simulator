# Negotiation Trainer API — MVP

## 1. О проекте

Сервис — B2B-тренажёр переговоров. Пользователь регистрируется, указывает одну организацию и может иметь одну или несколько продуктовых ролей. Для hackathon MVP намеренно не вводится сложная enterprise-модель membership/permissions: текущая схема с одной организацией и самостоятельным выбором ролей сохраняется.

Основные продуктовые роли:

- `learner` — проходит индивидуальные тренировки, участвует в Arena, смотрит историю и Profile;
- `methodist` — создаёт и публикует сценарии переговоров;
- `curator` — видит learner своей организации, назначает тренировки, планирует Arena и оставляет feedback.

Основной учебный цикл:

`Scenario -> Training/Arena -> Analysis/Feedback -> Profile`


## 3. Scenario

Scenario — опубликованный переговорный сценарий, который может использоваться многократно в Training и Arena.

### Жизненный цикл

- `draft` — редактируемый сценарий;
- `published` — доступен для Training/Arena и после публикации immutable;
- `archived` — не используется для новых запусков.

Версионирование в текущем MVP отсутствует. После публикации Scenario изменить нельзя.

### Основные свойства

Scenario содержит:

- `title` — название;
- `description` — исходное описание, введённое методистом;
- `expanded_context` — развёрнутый контекст; поле предусмотрено под AI-генерацию;
- `difficulty`: `easy | medium | hard`;
- `tags[]`;
- `supported_runtimes`: `deterministic | virtual_partner | arena`;
- `supported_channels`: `chat | email | voice_message | audio_call`;
- ровно две `roles`;
- `blocks` — состояния переговоров;
- `target_cluster_ids[]` — какие характеристики преимущественно тренирует Scenario.

AI-generation endpoints пока намеренно отсутствуют.

### Scenario roles

У каждой из двух ролей есть:

- `name`;
- `description`;
- `goal`;
- `interests`;
- `constraints`;
- `private_information`.

`interests`, `constraints` и `private_information` — свободный текст, а не структурированные массивы.

При Training/Arena пользователь получает `private_information` только своей роли.

### Blocks

Block — состояние переговоров, а не одна реплика. Внутри одного блока может пройти несколько обменов сообщениями.

Block содержит:

- `description`;
- `expanded_context`;
- `lead_role_id`;
- `lead_probability`;
- `role_goals[]`;
- `behavior`;
- `action_rules[]`;
- `constraints[]`;
- `deterministic_dialogue`;
- `transitions[]`.

`deterministic_dialogue` описывает многоходовое дерево реплик внутри блока. `transitions[]` отвечают только за выход из текущего блока в другой ScenarioBlock.

### Validation

`POST /scenarios/{scenario_id}/validate` возвращает `200` даже при `valid=false`.

Каждый issue структурирован:

```json
{
  "severity": "warning",
  "code": "BLOCK_LEAD_ROLE_MISSING",
  "message": "Состояние «Обсуждение цены» — не указано, кто ведущий",
  "entity_type": "block",
  "entity_id": "..."
}
```

Errors блокируют публикацию. Warnings не блокируют публикацию.

## 4. Кластеры и semantic actions

Справочники находятся в Scenarios API:

- `GET /competency-clusters`;
- `GET /semantic-actions`.

Кластеры не зашиваются enum-значениями в другие API. Profile, Training Analysis и Arena Feedback ссылаются на них через `cluster_id`.

Semantic action также передаётся стабильным `code` из справочника.

## 5. Каталог Scenario для learner

Отдельный TrainingCatalog resource не вводится. `GET /scenarios` одновременно является каталогом опубликованных тренировок для learner.

В `ScenarioSummary` добавлены персональные read-only поля:

- `is_favorite`;
- `is_assigned_by_curator`;
- `is_completed`;
- `attempts_count`;
- `best_attempt`.

Фильтры:

- `favorite`;
- `assigned_by_curator`;
- `completed`;
- `difficulty`;
- `tag`;
- `runtime`;
- `channel`.

Избранное:

- `PUT /scenarios/{scenario_id}/favorite`;
- `DELETE /scenarios/{scenario_id}/favorite`.

Один Scenario можно проходить неограниченное число раз.

В текущем контракте `best_attempt` — завершённая TrainingSession с максимальным `score` по шкале `0..100`. Если итоговая методика оценки изменится, правило определения best attempt можно поменять без изменения модели истории попыток.

## 6. Training

Старые основные endpoints сохранены:

- `POST /training/sessions`;
- `GET /training/sessions/{session_id}`;
- `POST /training/sessions/{session_id}/actions`;
- `POST /training/sessions/{session_id}/messages`;
- `POST /training/sessions/{session_id}/complete`;
- `GET /training/sessions/{session_id}/analysis`.

Добавлены:

- `GET /training/sessions` — история всех попыток текущего пользователя;
- `POST /training/sessions/{session_id}/messages/{message_id}/classification` — проверка semantic action входящей AI-реплики.

### Training modes

- `deterministic` — готовые реплики и переходы по Scenario;
- `virtual_partner` — свободный диалог с AI.

### Channels

Текущие Training channels:

- `chat`;
- `email`;
- `voice_message`.

`audio_call` пока не реализован Training API и добавится позже отдельным расширением enum/runtime.

### Semantic action собственных сообщений

Пользователь всегда размечает собственную реплику.

#### Chat

Одно сообщение -> один `semantic_action_code`.

#### Voice message

Одно голосовое сообщение -> один `semantic_action_code`.

Аудио сначала загружается через `POST /media`, затем `media_id` передаётся в Training message.

#### Email

Email содержит:

- `subject` — название/тема письма;
- `body` — текст письма;
- `semantic_annotations[]`.

Пользователь выделяет диапазоны в `body`, поэтому один email может иметь несколько semantic actions:

```json
{
  "start_offset": 15,
  "end_offset": 52,
  "semantic_action_code": "ASK_QUESTION"
}
```

Offsets относятся к `body`: start включительно, end не включительно.

### Проверка входящей AI-реплики

Ожидаемый semantic action AI-сообщения не возвращается вместе с сообщением.

Пользователь вызывает:

`POST /training/sessions/{session_id}/messages/{message_id}/classification`

и передаёт выбранный `semantic_action_code`. После этого backend возвращает:

- выбранный код;
- ожидаемый код;
- `correct`.

В текущем контракте одно AI-сообщение пользователь классифицирует один раз.

### Resume

Незавершённая попытка имеет статус `in_progress`. После закрытия страницы frontend может восстановить её через `GET /training/sessions/{session_id}`.

### Training analysis

После завершения доступны:

- итоговый `score`;
- фактический `result/outcome`;
- разбор `process`;
- `strengths[]`;
- `improvements[]`;
- оценки кластеров в конкретной попытке;
- `profile_change` с before/after/delta.

Evidence endpoint намеренно не добавлен на этом этапе.

## 7. Profile

Текущий Profile — набор значений кластеров `0..100`.

Endpoints:

- `GET /users/me/profile`;
- `GET /users/{user_id}/profile`;
- `GET /users/{user_id}/profile/history`.

История Profile хранит события изменения, а не только одну пару before/after.

Источник изменения:

- `training`;
- `arena_feedback`.

В текущем API нет отдельной ручки для evidence/observations.

## 8. Curator assignments

Старый `POST /assignments` сохранён.

Добавлен `GET /assignments`.

Assignment содержит только:

- curator;
- learner;
- Scenario;
- `pending | completed`.

Дедлайн отсутствует. Назначение не ограничивает каталог: learner по-прежнему может проходить любой опубликованный Scenario.

## 9. Curator team

Для MVP отдельная сущность Team отсутствует.

Команда curator = все пользователи с ролью `learner` в той же организации.

Endpoints:

- `GET /curator/team`;
- `GET /curator/team/sessions`;
- `GET /curator/team/analytics`.

`/curator/team/sessions` — read-model над TrainingSession и ArenaSession, а не отдельный источник истины.

## 10. Arena

Ключевой старый lifecycle сохранён:

`ArenaAssignment -> ArenaSession`

ArenaAssignment — запланированное событие.
ArenaSession — фактически начавшиеся переговоры.

Старые REST endpoints сохранены:

- `GET/POST /arena/assignments`;
- `GET /arena/assignments/{arena_assignment_id}`;
- `POST /arena/assignments/{arena_assignment_id}/join`;
- `GET /arena/sessions/{arena_session_id}`;
- `GET /arena/sessions/{arena_session_id}/messages`;
- `POST /arena/sessions/{arena_session_id}/complete`;
- `GET/POST /arena/sessions/{arena_session_id}/feedback`;
- `GET /users/{user_id}/arena-assignments`;
- `GET /users/{user_id}/arena-history`;
- `GET /users/{user_id}/arena-history/{arena_session_id}`.

Scenario version удалён: опубликованный Scenario immutable.

Participant теперь содержит `scenario_role_id`, а не старый enum `learner/partner`.

### Arena visibility

Рабочее правило MVP:

- все curator организации могут видеть Arena организации через `GET /arena/assignments`;
- feedback может создать только curator, который создал соответствующий ArenaAssignment.

### Feedback

Feedback:

- относится к конкретной ArenaSession;
- относится к одному participant;
- создаётся один раз;
- не редактируется.

`ratings[]` опциональны. В текущем контракте rating связан с `cluster_id` и использует шкалу `0..100`. Если методика разрешает, backend может преобразовать ratings в `profile_change`; иначе `profile_change=null`.

## 11. Arena WebSocket

Путь сохранён:

`/ws/arena/sessions/{arena_session_id}`

Client -> server:

- `message.send`.

Server -> client:

- `message.created`;
- `participant.joined`;
- `participant.left`;
- `arena.completed`;
- `error`.

`message_id` генерируется frontend и одновременно является постоянным ID и ключом идемпотентности.

Расширение старого message payload:

```json
{
  "type": "message.send",
  "message_id": "...",
  "text": "Что для вас важно?",
  "semantic_action_code": "ASK_QUESTION"
}
```

REST `GET /arena/sessions/{arena_session_id}/messages` остаётся источником восстановления истории после reconnect.

## 12. Что намеренно отложено

В текущую спецификацию не входят:

- AI Scenario Builder endpoints;
- provenance AI generations;
- evidence/observations API;
- `annotation_mode`;
- Scenario versioning;
- редактирование опубликованного Scenario;
- отдельная сущность curator Team;
- deadlines для Training Assignment;
- audio_call runtime;
- realtime voice protocol;
- редактирование Arena feedback;
- сложная B2B membership/admin model.

Эти расширения можно добавить позднее без необходимости менять основные сущности TrainingSession, ArenaAssignment/ArenaSession и Profile.