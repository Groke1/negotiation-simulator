# AGENTS.md — Negotiation Trainer API MVP

## 1. Назначение файла

Этот файл задаёт рабочий контекст для coding-agent, который изменяет backend, OpenAPI/AsyncAPI, модели и runtime сервиса Negotiation Trainer.

Главная цель MVP — показать работающий тренажёр переговоров. Не усложнять решение enterprise-механиками, которые не нужны для демонстрации основного учебного цикла.

Основной цикл продукта:

`Scenario -> Training/Arena -> Analysis/Feedback -> Profile`

При изменениях API и кода сначала сохранять уже принятые контракты и доменные инварианты. Новые сущности и endpoints добавлять только когда существующую модель нельзя разумно расширить.

## 2. Источники истины и правила изменений

Для конкретных HTTP/WS payload, обязательности полей, кодов ошибок и status code источником истины являются текущие OpenAPI/AsyncAPI-файлы репозитория.

Этот `AGENTS.md` фиксирует продуктовую модель и архитектурные решения. Если комментарий в коде или старое описание противоречит текущему OpenAPI, не восстанавливать старое поведение автоматически.

При изменении API:

- по возможности сохранять существующие URL и operation semantics;
- не переименовывать уже используемые поля без необходимости;
- расширять существующие schemas дополнительными полями вместо создания параллельных ресурсов;
- возвращать структурированные ошибки в формате текущего API (`code`, `message` и дополнительные поля конкретной ошибки, если они предусмотрены схемой);
- не добавлять speculative/future функциональность только ради универсальности архитектуры.

Текущая логическая раскладка спецификаций:

- `scenarios.yaml` — Scenario, конструктор, validation, справочники кластеров и semantic actions, каталог Scenario;
- `training.yaml` — TrainingSession, deterministic runtime, virtual partner, history, analysis;
- `users.yaml` — пользователь, роли/организация и Profile endpoints;
- `curator.yaml` — assignments, curator team, team sessions, analytics;
- `arena.yaml` — Arena REST;
- `arena_asyncapi.yaml` — Arena WebSocket;
- `media.yaml` — загрузка и получение media;
- auth/organization/internal APIs — существующий MVP onboarding и служебные операции.


## 3. Границы MVP

Это hackathon MVP. Намеренно сохраняется простая B2B-модель:

- пользователь сам регистрируется;
- пользователь указывает одну организацию;
- пользователь может иметь несколько продуктовых ролей;
- пользователь сам выбирает/добавляет роли согласно текущему MVP API;
- сложная membership/admin/permission model не нужна;
- отдельная сущность curator Team не нужна.

Не проектировать enterprise RBAC, invitation flow, multi-organization membership и org admin, пока это явно не потребуется.

## 4. Продуктовые роли

### learner

Обучающийся пользователь. Может:

- просматривать каталог опубликованных Scenario;
- добавлять Scenario в избранное;
- проходить Scenario многократно;
- проходить назначенные curator Scenario;
- использовать deterministic и virtual-partner Training;
- смотреть историю попыток и их analysis;
- смотреть свой Profile и его историю;
- видеть назначенные Arena;
- участвовать в Arena;
- смотреть историю Arena и feedback.

### methodist

Создаёт переговорные Scenario:

- создаёт draft;
- задаёт две роли;
- строит блоки и переходы;
- задаёт deterministic dialogue;
- запускает validation;
- публикует Scenario;
- архивирует Scenario.

AI Scenario Builder в текущем API отсутствует, хотя поля `expanded_context` заранее предусмотрены под будущую AI-генерацию.

### curator

Работает с learner своей организации:

- видит всех пользователей организации с ролью `learner`;
- смотрит их Profile и историю;
- смотрит Training/Arena activity;
- назначает Scenario как Training Assignment;
- создаёт ArenaAssignment для двух пользователей;
- видит Arena организации;
- оставляет feedback по созданной им Arena;
- смотрит командную аналитику.

## 5. Справочник кластеров

Кластеры — характеристики/компетенции Profile пользователя.

Справочник отдаётся через:

`GET /competency-clusters`

Не зашивать список кластеров enum-полями в Profile, Analysis или Feedback. Эти модели ссылаются на справочник через `cluster_id`.

Для текущего negotiation domain использовать названия из продуктового прототипа:

| Prototype key | Название |
| --- | --- |
| `goals` | Цели и границы |
| `diag` | Диагностика |
| `value` | Создание ценности |
| `trade` | Торг и обмен |
| `adapt` | Адаптивность |
| `close` | Завершение |

`Prototype key` — удобный стабильный seed/reference key, но внешние модели должны по возможности использовать `cluster_id`, а не enum этих значений.

Profile использует шкалу `0..100`.

## 6. Semantic actions / метки сообщений

Справочник semantic actions отдаётся через:

`GET /semantic-actions`

Human-readable названия для negotiation domain брать из прототипа:

- **Задать вопрос**;
- **Уточнить интерес**;
- **Предложить вариант**;
- **Уступить**;
- **Запросить взаимность**;
- **Удержать границу**;
- **Завершить переговоры**.

В прототипе также есть системная метка:

- **Действие отсутствует** — в реплике нет признаков ни одного действия из словаря.

Коды semantic actions являются стабильными backend codes из справочника. Не вычислять code из русской подписи на клиенте и не считать label идентификатором. Если в текущем OpenAPI уже существуют стабильные codes, сохранять их.

## 7. Scenario

Scenario описывает переговорную ситуацию и после публикации может многократно использоваться в Training и Arena.

### Lifecycle

Статусы:

- `draft` — редактируемый;
- `published` — доступен для запусков и immutable;
- `archived` — не используется для новых запусков.

Scenario versioning в MVP отсутствует.

После `published` содержимое Scenario не редактируется. Для изменения опубликованного сценария в будущем можно создать новый Scenario; не вводить скрытое versioning без отдельного решения.

### Основные поля

Scenario содержит как минимум:

- `id`;
- `title`;
- `description` — исходное описание методиста;
- `expanded_context` — будущий/сгенерированный расширенный контекст;
- `difficulty`: `easy | medium | hard`;
- `tags[]`;
- `supported_runtimes`: `deterministic | virtual_partner | arena`;
- `supported_channels`: `chat | email | voice_message | audio_call` на уровне возможностей Scenario;
- ровно две `roles`;
- `blocks`;
- `target_cluster_ids[]`;
- lifecycle timestamps/status согласно OpenAPI.


### Roles

В Scenario ровно две переговорные роли.

Роль содержит:

- `id`;
- `name`;
- `description`;
- `goal`;
- `interests`;
- `constraints`;
- `private_information`.

`interests`, `constraints`, `private_information` — свободный текст, а не массивы структур.

Не путать product role (`learner`, `methodist`, `curator`) с ролью переговоров внутри Scenario (`Закупщик`, `Поставщик` и т.п.).

При Training/Arena участник получает private information только своей ScenarioRole. Backend не должен раскрывать private information другой стороны.

## 8. ScenarioBlock и граф переговоров

Block — **состояние переговоров**, а не отдельная реплика.

Внутри одного блока может происходить несколько последовательных обменов сообщениями.

Block содержит концептуально:

- `id`;
- type/start-end metadata согласно OpenAPI;
- `name`;
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

Отдельного поля `context` в Block не вводить.

`role_goals[]` задаёт цель каждой ScenarioRole внутри конкретного состояния.

### Межблочные transitions

`transitions[]` описывают только выход из одного ScenarioBlock в другой.

Transition обычно содержит:

- `id`;
- `target_block_id`;
- `condition`;
- optional методическую заметку;
- `weight`;
- `forbidden`.

Не использовать межблочный transition как следующий ход внутри многоходового deterministic dialogue.

## 9. Deterministic dialogue внутри Block

`deterministic_dialogue` — отдельное дерево реплик **внутри одного Block**.

Структура:

- `start_node_id` — первая dialogue node;
- `nodes[]` — возможные состояния детерминированного разговора.

Каждая node содержит:

- `id`;
- `incoming_message` — заранее подготовленную реплику второй стороны;
- `reply_options[]` — варианты ответа пользователя.

Каждый reply option ведёт одним из двух способов:

1. `next_node_id` — продолжить диалог внутри того же ScenarioBlock;
2. `transition_id` — закончить текущую внутриблочную ветку и применить переход в другой ScenarioBlock.

Логически у reply option должен использоваться ровно один способ продолжения: `next_node_id` **или** `transition_id`.

Пример модели исполнения:

`start node -> incoming message -> learner reply option -> next node -> incoming message -> ... -> transition -> next ScenarioBlock`

Дерево может ветвиться и при необходимости содержать циклы, если это разрешает validation/runtime.

Scenario хранит правильную semantic-разметку deterministic content. Во время Training пользователь отдельно отправляет выбранную им semantic action; backend может сравнивать пользовательскую классификацию с эталонной.

## 10. Scenario validation

Endpoint:

`POST /scenarios/{scenario_id}/validate`

`valid=false` — нормальный результат бизнес-валидации и возвращается с HTTP `200`.

Issue должен быть структурированным, чтобы frontend мог подсветить сущность/поле, например:

```json
{
  "severity": "warning",
  "code": "BLOCK_LEAD_ROLE_MISSING",
  "message": "Состояние «Обсуждение цены» — не указано, кто ведущий",
  "entity_type": "block",
  "entity_id": "..."
}
```

Errors блокируют публикацию. Warnings не блокируют.

Минимальные проверки графа:

- существует корректный start;
- существует end;
- блоки достижимы от start;
- у обычного блока есть путь к завершению;
- отсутствующие исходящие transitions дают warning, если это не допустимый end;
- блок, у которого все transitions forbidden, является ошибкой;
- lead role должна ссылаться на существующую роль;
- transition target должен существовать;
- пустые/подозрительно дублирующиеся conditions отмечаются;
- ссылки deterministic dialogue (`start_node_id`, `next_node_id`, `transition_id`) должны быть валидны.

Publish должен повторно проверять Scenario и не публиковать draft с validation errors.

## 11. Каталог Scenario для learner

Отдельного TrainingCatalog resource нет.

`GET /scenarios` используется и как список Scenario, и как learner catalog.

Для learner summary может содержать персональные read-only поля:

- `is_favorite`;
- `is_assigned_by_curator`;
- `is_completed`;
- `attempts_count`;
- `best_attempt`.

Поддерживаются фильтры из текущего OpenAPI, включая:

- favorite;
- assigned by curator;
- completed;
- difficulty;
- tag;
- runtime;
- channel.

Favorites:

- `PUT /scenarios/{scenario_id}/favorite`;
- `DELETE /scenarios/{scenario_id}/favorite`.

Один опубликованный Scenario можно проходить многократно.

`best_attempt` в текущем MVP — завершённая TrainingSession с максимальным итоговым `score` по шкале `0..100`.

## 12. TrainingSession

TrainingSession — одна конкретная попытка одного learner пройти Scenario.

Один Scenario может иметь любое количество TrainingSession одного пользователя.

Сохранять существующие основные endpoints:

- `POST /training/sessions`;
- `GET /training/sessions`;
- `GET /training/sessions/{session_id}`;
- `POST /training/sessions/{session_id}/actions`;
- `POST /training/sessions/{session_id}/messages`;
- `POST /training/sessions/{session_id}/complete`;
- `GET /training/sessions/{session_id}/analysis`;
- `POST /training/sessions/{session_id}/messages/{message_id}/classification`.

Не создавать отдельные endpoints на каждый channel, если существующий message endpoint можно расширить discriminated schema.

### Modes

- `deterministic` — пользователь выбирает подготовленные варианты внутри deterministic dialogue;
- `virtual_partner` — пользователь свободно общается с AI-партнёром.

### Channels

Фактически реализуемые Training channels MVP:

- `chat`;
- `email`;
- `voice_message`.

`audio_call` может быть указан как будущая возможность Scenario, но Training audio-call runtime и realtime voice protocol сейчас не реализуются.

### Resume/status

Незавершённая попытка сохраняется с `in_progress` и должна восстанавливаться через `GET /training/sessions/{session_id}`.

Не считать закрытие браузера завершением Training.

## 13. Message semantic annotation

Пользователь размечает свои сообщения сам.

### Chat

Одно chat message имеет один:

`semantic_action_code`

### Voice message

Одно voice message имеет один:

`semantic_action_code`

Voice message ссылается на `media_id`.

### Email

Email содержит:

- `subject` — тема письма;
- `body` — тело письма;
- `semantic_annotations[]`.

Один email может содержать несколько semantic actions. Пользователь выделяет конкретные части `body`.

Annotation:

```json
{
  "start_offset": 15,
  "end_offset": 52,
  "semantic_action_code": "..."
}
```

Offsets относятся только к `body`:

- `start_offset` inclusive;
- `end_offset` exclusive.

Не применять один глобальный semantic action ко всему email, если пользователь разметил отдельные диапазоны.

### Классификация входящего AI message

Ожидаемый semantic action AI-сообщения нельзя отдавать клиенту до ответа пользователя.

Пользователь отправляет свой выбор через:

`POST /training/sessions/{session_id}/messages/{message_id}/classification`

После этого backend возвращает выбранный код, эталонный/ожидаемый код и `correct`.

В текущем MVP одно AI message классифицируется пользователем один раз.

## 14. Media

Media нужен прежде всего для `voice_message`.

Основной поток:

1. клиент загружает media через `POST /media`;
2. backend возвращает идентификатор/метаданные media согласно текущему OpenAPI;
3. client передаёт `media_id` в Training message;
4. для воспроизведения клиент получает сам media-файл через backend Media GET endpoint.

**Не возвращать и не хранить в API публичный `url` для media.**

Media извлекается через backend (текущий ресурс `/media/{media_id}`), с обычной проверкой авторизации/доступа. GET должен отдавать содержимое media согласно актуальному `media.yaml`, а не ссылку на object storage.

Для MVP этого достаточно. Presigned URLs, streaming protocol и realtime audio не нужны.

## 15. Training Analysis

После завершения Training доступен analysis.

Он должен уметь вернуть:

- итоговый `score` (`0..100`);
- фактический `result/outcome`;
- разбор `process`;
- `strengths[]`;
- `improvements[]`;
- оценки кластеров для конкретной попытки;
- `profile_change` с `before`, `after`, `delta`.

Outcome и competency/profile — разные понятия. Удачная сделка сама по себе не означает высокую переговорную компетентность.

Evidence/observations endpoint сейчас намеренно отсутствует. Не добавлять его без отдельного продуктового решения.

## 16. Profile

Profile — текущее агрегированное состояние переговорных характеристик learner.

Шкала кластеров: `0..100`.

Profile endpoints находятся в `users.yaml`:

- `GET /users/me/profile`;
- `GET /users/{user_id}/profile`;
- `GET /users/{user_id}/profile/history`.

`GET /users/{user_id}/profile` доступен самому пользователю и разрешённому curator той же организации согласно текущей спецификации.

Profile history — последовательность событий изменения, а не только последняя пара before/after.

Возможные источники изменения:

- `training`;
- `arena_feedback`.

Не добавлять отдельный evidence/observations API на текущем этапе.

## 17. Training Assignment

Curator может назначить опубликованный Scenario learner.

Сохранять существующий resource `/assignments`:

- `POST /assignments`;
- `GET /assignments`.

Assignment связывает:

- curator;
- learner;
- Scenario;
- status `pending | completed`.

Deadline в MVP отсутствует.

Assignment не ограничивает доступ к каталогу. Learner всё равно может самостоятельно запускать любой опубликованный Scenario.

Завершённый назначенный Scenario остаётся доступным для повторных прохождений.

## 18. Curator team и analytics

Отдельной сущности Team нет.

Команда curator = все пользователи с ролью `learner` в той же организации.

Основные endpoints:

- `GET /curator/team`;
- `GET /curator/team/sessions`;
- `GET /curator/team/analytics`.

`/curator/team/sessions` — read-model над существующими TrainingSession и ArenaSession. Не создавать отдельную persisted session-history сущность только для dashboard.

Profile endpoints не должны дублироваться в `curator.yaml`; они находятся в `users.yaml` и защищаются authorization rules.

## 19. Arena

Arena — переговоры двух реальных пользователей по опубликованному Scenario.

Не путать:

`ArenaAssignment` — запланированное событие.

`ArenaSession` — фактически начавшиеся переговоры.

Сохранять существующий lifecycle:

`ArenaAssignment -> ArenaSession`

Сохранять существующие REST paths, если нет явной причины менять контракт:

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

Scenario version в Arena отсутствует, потому что published Scenario immutable.

Arena participant содержит `scenario_role_id`, а не enum `learner/partner`.

Каждый участник получает private information только своей ScenarioRole.

### Visibility

MVP rule:

- curator организации может видеть Arena своей организации через `/arena/assignments`;
- feedback может создать только curator, создавший соответствующий ArenaAssignment.

## 20. Arena feedback

Feedback:

- относится к одной ArenaSession;
- относится к одному participant;
- создаётся curator;
- создаётся один раз;
- после создания не редактируется.

Повторное создание feedback для того же participant/session должно обрабатываться как конфликт согласно OpenAPI.

`ratings[]` optional.

Rating:

- ссылается на `cluster_id`;
- использует шкалу `0..100` в текущем MVP;
- оценивает выступление в конкретной Arena, а не напрямую устанавливает значение Profile.

Backend может преобразовать feedback/rating в изменение Profile, если текущая методика это разрешает. В противном случае `profile_change` может отсутствовать/null согласно контракту.

## 21. Arena WebSocket

Максимально сохранять старый WebSocket protocol.

Путь:

`/ws/arena/sessions/{arena_session_id}`

Client -> server:

- `message.send`.

Server -> client:

- `message.created`;
- `participant.joined`;
- `participant.left`;
- `arena.completed`;
- `error`.

Frontend генерирует `message_id` до отправки.

`message_id` является одновременно:

- постоянным ID сообщения;
- idempotency key для retry/reconnect.

Повторный `message.send` с тем же `message_id` не должен создавать дубликат.

Arena chat message расширен одним `semantic_action_code`, выбранным отправителем.

REST:

`GET /arena/sessions/{arena_session_id}/messages`

остаётся источником восстановления истории после reconnect и после завершения WebSocket connection.

Не переносить историю Arena исключительно в ephemeral WebSocket state.

## 22. Ошибки и HTTP semantics

Следовать текущим OpenAPI schemas ошибок.

Базовая форма:

```json
{
  "code": "SOME_ERROR_CODE",
  "message": "Human-readable message"
}
```

Типично используются:

- `400` — malformed/invalid request;
- `401` — нет валидной auth session/token;
- `403` — роль/доступ запрещены;
- `404` — resource не найден или недоступен согласно выбранной security semantics;
- `409` — конфликт состояния/idempotency/already exists;
- `422` — корректный по форме запрос нельзя выполнить из-за бизнес-валидации, если именно так задан текущий endpoint;
- `500` — внутренняя ошибка.

Не заменять бизнес-результат Scenario validation на HTTP error: `valid=false` возвращается через успешный validation response.

## 23. Что намеренно не реализуем сейчас

Без отдельного запроса не добавлять:

- AI Scenario Builder endpoints;
- provenance/history AI generations;
- evidence/observations API;
- `annotation_mode`;
- Scenario versioning;
- редактирование published Scenario;
- отдельную persisted curator Team;
- deadlines для Training Assignment;
- working `audio_call` runtime;
- realtime voice protocol;
- редактирование Arena feedback;
- multi-organization membership;
- invitation/admin enterprise flow;
- сложную систему permissions;
- presigned/public media URLs.

## 24. Ключевые доменные инварианты

При любых изменениях сохранять следующее:

1. Product role и ScenarioRole — разные понятия.
2. ScenarioBlock — состояние переговоров, не отдельное сообщение.
3. Deterministic dialogue nodes живут внутри Block; transitions между blocks — отдельный уровень графа.
4. Published Scenario immutable; versioning отсутствует.
5. У Scenario ровно две переговорные роли.
6. Private information одной роли не раскрывается другой стороне.
7. Один Scenario можно проходить многократно; каждая попытка — отдельная TrainingSession.
8. Assignment не блокирует самостоятельное прохождение Scenario.
9. Chat и voice message имеют одну пользовательскую semantic action; email может иметь несколько annotations по диапазонам body.
10. Правильная semantic action входящего AI message не раскрывается до попытки классификации пользователя.
11. Profile и outcome не взаимозаменяемы.
12. ArenaAssignment и ArenaSession — разные сущности.
13. Arena всегда предполагает двух реальных участников.
14. Arena messages сохраняются независимо от WebSocket connection.
15. Client-generated `message_id` обеспечивает idempotency Arena chat.
16. Arena feedback immutable после создания.
17. Profile endpoints находятся в `users.yaml`.
18. Media API не возвращает публичный URL; media читается через backend.

## 25. Checklist для coding-agent

Перед изменением модели/API проверить:

- можно ли расширить существующий endpoint вместо создания нового;
- не раскрывается ли чужая `private_information`;
- не смешаны ли product roles и Scenario roles;
- не смешаны ли dialogue node и ScenarioBlock transition;
- не ломается ли resume/idempotency;
- не появляется ли скрытое Scenario versioning;
- не захардкожены ли кластеры вместо `cluster_id` там, где нужен справочник;
- semantic action берётся ли из справочника, а не из UI label;
- учтена ли особая разметка email ranges;
- не возвращается ли media URL;
- Profile endpoints изменяются ли в `users.yaml`, а не дублируются в curator API;
- соответствуют ли ошибки текущим OpenAPI error schemas;
- не реализуется ли отложенная функциональность без отдельного требования.