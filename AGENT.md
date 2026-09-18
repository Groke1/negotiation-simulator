# Арена переговоров — предметная модель проекта

Этот файл описывает предметную область проекта для AI-агента. Здесь зафиксированы основные сущности, роли пользователей, связи между ними и актуальные продуктовые возможности. Технический стек, структура пакетов и детали реализации намеренно не описываются, если они не важны для понимания домена.

## 1. Общая идея продукта

«Арена переговоров» — система для тренировки переговорных навыков.

Основной учебный цикл:

**Диагностика → Профиль → Рекомендация → Практика → обновление Профиля**

В продукте есть два принципиально разных типа практики:

1. **Training** — индивидуальная тренировка пользователя.
2. **Arena** — запланированные переговоры двух реальных пользователей.

Training, в свою очередь, может работать в двух runtime-режимах:

- `deterministic` — пользователь выбирает готовые действия в заранее заданном графе;
- `virtual_partner` — пользователь свободно переписывается с виртуальным партнёром.

В текущем MVP основной рабочий канал — `chat`.

Голосовой режим и аудиозаписи рассматриваются как дальнейшее развитие и не должны усложнять текущую chat-only модель.

## 2. User

`User` — аккаунт человека в системе.

У пользователя есть:

- username;
- email;
- пароль;
- набор назначенных продуктовых ролей;
- одна активная роль в текущей auth session;
- одна организация в рамках текущего MVP.

После регистрации пользователь по умолчанию получает роль `learner`.

Один пользователь может иметь несколько ролей одновременно, например `learner`, `methodist` и `curator`, но в конкретной auth session активна только одна роль. Активная роль определяет доступный интерфейс и разрешённые действия.

## 3. Organization

`Organization` — компания или другая организационная область, внутри которой работают пользователи.

В текущем MVP пользователь может состоять только в одной организации.

Scenario относятся к конкретной организации. Curator работает с пользователями своей организационной области.

Целевая модель должна оставаться расширяемой на случай, если позже одному пользователю разрешат состоять в нескольких организациях.

## 4. Продуктовые роли

### learner

Обучающийся.

Может:

- проходить Training;
- смотреть свой Profile;
- смотреть историю Training;
- смотреть разбор завершённых Training;
- получать Recommendation;
- выполнять назначенные curator тренировки;
- видеть запланированные Arena;
- подключаться к назначенной Arena;
- смотреть историю своих Arena;
- перечитывать сохранённый диалог Arena;
- смотреть экспертный feedback по завершённой Arena.

### methodist

Методист / разработчик курсов.

Может:

- создавать Scenario вручную;
- генерировать draft через AI Scenario Builder;
- редактировать Scenario;
- запускать validation;
- публиковать Scenario;
- архивировать Scenario;
- смотреть опубликованные версии;
- проходить preview Training перед публикацией.

ИИ помогает methodist создать draft, но решение о публикации принимает человек.

### curator

Куратор / эксперт.

Может:

- видеть свою команду;
- видеть активность пользователей;
- смотреть Profile сотрудников;
- смотреть историю Training;
- смотреть динамику Profile;
- назначать опубликованные Training;
- видеть командную аналитику;
- выбирать двух пользователей для Arena;
- выбирать Scenario для Arena;
- задавать дату и время Arena;
- смотреть завершённый диалог Arena;
- оставлять отдельный feedback каждому участнику Arena.

В текущей модели отдельная продуктовая роль `expert` не вводится. Экспертный review Arena выполняет пользователь с ролью `curator`.

Curator не редактирует Scenario.

## 5. Scenario

`Scenario` — описание переговорного мира.

Это не конкретное прохождение. Один Scenario может использоваться много раз и в разных runtime.

Scenario включает:

- название;
- версию;
- поддерживаемые runtime;
- context;
- роли сторон переговоров;
- goals;
- interests;
- constraints;
- parameters;
- private information;
- difficulty;
- competencies;
- observation rules;
- completion conditions;
- channels;
- `nodes` для deterministic Training;
- `partner_behavior` для virtual-partner Training.

Scenario может поддерживать один или несколько runtime:

- `deterministic`;
- `virtual_partner`;
- `arena`.

Не все поля Scenario обязательны для каждого runtime.

Например:

- `nodes` нужны для `deterministic`;
- `partner_behavior` нужен для `virtual_partner`;
- для `arena` переговорную динамику создают два реальных человека, поэтому граф и virtual partner не используются как runtime-механика.

## 6. Роли внутри Scenario

Не путать с продуктовыми ролями пользователя.

Внутри переговоров используются роли сторон:

- `learner`;
- `partner`.

Например:

- learner — «Менеджер по продажам»;
- partner — «Клиент».

В Arena эти две роли распределяются между двумя реальными пользователями.

## 7. Context, goals, interests и constraints

`context` задаёт исходную ситуацию: тип переговоров, тип конфликта и краткое описание.

`goals` описывает, чего каждая сторона хочет добиться.

`interests` описывает причины, стоящие за позициями сторон.

`constraints` содержит жёсткие ограничения, например:

- максимальную скидку;
- минимально приемлемый срок;
- обязательные условия сделки;
- наличие альтернативного поставщика.

Эти поля используются и в Training, и в Arena.

## 8. Parameters

`parameters` — изменяемые параметры сделки, например:

- цена;
- размер скидки;
- срок контракта;
- объём;
- SLA;
- распределение рисков.

Они позволяют фиксировать фактический результат переговоров отдельно от оценки поведения.

## 9. Private information

`private_information` — информация, известная только одной стороне.

Например, learner знает, что может предложить расширенный SLA, а partner знает, что у него есть альтернативный поставщик.

Private information особенно важно для:

- virtual-partner Training;
- Arena двух реальных пользователей.

Стороны не должны автоматически получать private information друг друга.

## 10. Difficulty

В текущей модели используются уровни:

- `L1`;
- `L2`;
- `L3`.

Сложность задаёт уровень давления, неоднозначности и требований к поведению обучающегося.

## 11. Competencies

Используются шесть основных кластеров:

1. Цели и границы
2. Диагностика
3. Создание ценности
4. Аргументация
5. Торг и обмен
6. Конфликт и регуляция

Адаптивность не является отдельной седьмой компетенцией. Она рассматривается как устойчивость и изменение поведения при росте сложности и давления.

## 12. Semantic actions

Действия пользователя могут быть представлены семантическими типами, независимыми от конкретной формулировки.

Основные semantic actions:

- `ASK_QUESTION`
- `MAKE_PROPOSAL`
- `MAKE_CONCESSION`
- `REQUEST_RECIPROCITY`
- `HOLD_BOUNDARY`
- `CHANGE_TOPIC`
- `PAUSE`
- `SUMMARIZE`
- `CLOSE_NEGOTIATION`

Для deterministic Training semantic action известен заранее, потому что он привязан к выбранному action.

Для virtual-partner Training пользователь пишет свободный текст. В этом случае semantic actions могут извлекаться отдельным semantic classifier.

Одна свободная реплика может содержать несколько semantic actions.

LLM не должна напрямую выставлять competency score только на основании субъективной оценки текста. Предпочтительная схема:

**текст пользователя → semantic classification → observation rules → competency observations**

## 13. Deterministic graph

Граф используется только в `deterministic` Training.

Scenario содержит `nodes`.

Node описывает:

- идентификатор;
- тип узла;
- говорящую сторону;
- текст реплики;
- возможную эмоцию/давление;
- доступные действия пользователя;
- переходы к следующим узлам;
- при необходимости terminal outcome.

Основные типы узлов:

- `opening`;
- `escalation`;
- `resolution`.

Каждое доступное действие пользователя связано с `semantic_action`.

Граф не используется для Arena двух реальных пользователей.

## 14. Virtual partner behavior

Для Training в режиме `virtual_partner` вместо детерминированного графа используется описание поведения виртуального партнёра.

`partner_behavior` может задавать:

- persona;
- communication style;
- initial position;
- pressure policy;
- information policy;
- concession policy;
- forbidden behavior.

Virtual partner должен действовать внутри целей, интересов, constraints и private information Scenario.

Runtime virtual partner не должен самостоятельно менять смысл Scenario или нарушать жёсткие ограничения роли.

Состояние переговоров желательно хранить структурированно на стороне backend, а не только в контексте LLM.

## 15. Observation rules

`observation_rules` описывают, как наблюдаемые действия пользователя превращаются в наблюдения по компетенциям.

Правило связывает:

- semantic action;
- competency cluster;
- конкретный indicator;
- условие, при котором действие считается проявлением индикатора.

Пример:

`REQUEST_RECIPROCITY` → «Торг и обмен» → «Запрашивает встречную ценность».

Для deterministic Training semantic action приходит непосредственно из выбранного action.

Для virtual-partner Training semantic action сначала извлекается из свободного текста.

Scoring Training должен оставаться максимально объяснимым и опираться на observation rules, а не на непрозрачную итоговую оценку LLM.

## 16. Completion conditions

`completion_conditions` определяют, когда Training считается завершённым.

Например:

- достигнуто соглашение;
- переговоры завершены без сделки;
- достигнуты нужные значения parameters;
- learner явно завершил переговоры.

Для Arena завершение определяется жизненным циклом Arena Session, а не прохождением графа.

## 17. Channels

В текущем MVP используется:

- `chat`.

В дальнейшем могут появиться:

- `voice`;
- аудиозаписи;
- другие real-time форматы.

Пока voice не должен считаться обязательным рабочим каналом.

## 18. Статусы Scenario

Используются:

- `draft`;
- `published`;
- `archived`.

`draft` — редактируемый черновик, созданный вручную или через AI Scenario Builder.

`published` — Scenario доступен для использования. При публикации создаётся immutable ScenarioVersion.

`archived` — Scenario больше не используется для новых назначений/прохождений, но старые версии и история сохраняются.

## 19. ScenarioVersion

`ScenarioVersion` — неизменяемый опубликованный snapshot Scenario.

Он нужен, чтобы старые Training и Arena всегда ссылались на ту версию, которая реально использовалась.

Изменение текущего Scenario не должно менять историю уже завершённых прохождений.

Важно различать:

- `Scenario` — редактируемая сущность;
- `ScenarioVersion` — опубликованный immutable snapshot.

## 20. Validation Scenario

Методист может отдельно проверить Scenario перед публикацией.

Для deterministic runtime проверяются:

- базовая полнота структуры;
- тупики графа;
- недостижимые узлы;
- циклы;
- корректность переходов;
- покрытие заявленных indicators и competencies.

Для virtual-partner runtime дополнительно проверяется наличие достаточного описания `partner_behavior`.

Для Arena проверяется, что Scenario содержит достаточный контекст для двух реальных участников.

`valid = false` — нормальный результат проверки, а не техническая ошибка.

После изменения Scenario предыдущий validation считается устаревшим.

## 21. TrainingSession

`TrainingSession` — конкретное индивидуальное прохождение Scenario пользователем.

TrainingSession всегда относится к одному learner.

TrainingSession содержит:

- идентификатор;
- Scenario;
- конкретную ScenarioVersion;
- `mode`;
- `channel`;
- learner;
- runtime state;
- историю действий или сообщений;
- outcome;
- analysis;
- статус.

Допустимые Training mode:

- `deterministic`;
- `virtual_partner`.

### deterministic

Пользователь получает текущий node и список available actions.

После выбора action backend:

1. фиксирует действие;
2. знает semantic action;
3. применяет observation rules;
4. переводит Session в следующий node.

### virtual_partner

Пользователь пишет свободное сообщение.

Backend:

1. сохраняет сообщение learner;
2. определяет semantic actions;
3. применяет observation rules;
4. передаёт контекст virtual partner;
5. получает ответ virtual partner;
6. сохраняет ответ;
7. обновляет runtime state.

## 22. Training Event Log

Детерминированные действия Training записываются как последовательность событий.

Event может содержать:

- идентификатор;
- порядковый номер;
- actor;
- semantic action;
- node;
- выбранный action;
- время.

Для virtual-partner Training кроме event/observation history сохраняется история chat messages.

Event Log используется для воспроизводимого разбора Training.

## 23. Training chat messages

В `virtual_partner` Training сохраняются сообщения:

- learner;
- virtual partner.

Message содержит как минимум:

- `message_id`;
- author;
- text;
- created_at.

`message_id` генерируется frontend до отправки и сохраняется в базе как идентификатор сообщения. Повторная отправка сообщения с тем же `message_id` после retry/reconnect не создаёт дубликат.

## 24. Outcome

`outcome` описывает фактический результат переговоров.

Например:

- контракт продлён;
- цена изменилась;
- SLA изменён;
- достигнут определённый срок;
- отдельные цели достигнуты или не достигнуты.

Важно:

**Outcome не является оценкой компетентности.**

Выгодный результат сделки и качественный переговорный процесс — разные вещи.

## 25. Training Analysis

После завершения Training пользователь получает разбор.

Основные слои:

1. **Result** — что произошло с условиями сделки.
2. **Process** — какие действия совершал learner.
3. **Competency** — какие поведенческие паттерны проявились.

Для virtual-partner Training Process строится с учётом сохранённого диалога и semantic observations.

## 26. Profile

`Profile` — текущая модель переговорных компетенций пользователя.

Профиль состоит из шести competency clusters и связанных indicators.

Для indicator может храниться:

- уровень;
- источник наблюдения;
- время последнего обновления.

Уровни предполагаются в диапазоне `0..4`.

Profile обновляется после завершённых Training.

Arena пока не должна автоматически менять Profile, пока не определена методика экспертных оценок Arena.

## 27. История Profile

Необходимо различать текущее состояние Profile и историю наблюдений.

История нужна, чтобы:

- сравнивать состояние «до» и «после»;
- показывать динамику;
- давать curator картину развития сотрудника.

## 28. Recommendation

`Recommendation` — следующая рекомендуемая Training пользователя.

Источники рекомендации:

- автоматический recommendation engine;
- явное назначение curator.

Автоматическая рекомендация строится на основе слабых мест текущего Profile.

Активное назначение curator может иметь приоритет над автоматической рекомендацией.

Arena schedule является отдельным механизмом и не должна смешиваться с Recommendation Training.

## 29. Training Assignment

`Assignment` — назначение опубликованной Training curator конкретному learner.

Assignment связывает:

- curator;
- learner;
- опубликованный Scenario / ScenarioVersion;
- статус выполнения.

Основные статусы:

- `pending`;
- `completed`.

Assignment должен быть привязан к конкретной опубликованной версии.

Не путать Training Assignment с ArenaAssignment.

## 30. Команда curator

Curator работает с набором закреплённых learner.

Для участника команды могут использоваться:

- пользователь;
- активность;
- дата последней завершённой Training;
- самый слабый competency cluster;
- текущий Profile;
- история Training;
- динамика Profile;
- запланированные Arena;
- история Arena.

## 31. Командная аналитика

Curator может видеть агрегированную информацию по команде, например:

- распределение пользователей по уровням competency clusters;
- наиболее часто встречающийся слабый cluster;
- активность пользователей;
- прохождение назначенных Training;
- участие в Arena.

Командная аналитика является производной от пользовательских данных и не является отдельным независимым источником истины.

## 32. AI Scenario Builder

ИИ используется как помощник methodist.

Methodist вводит естественноязычное описание, после чего система создаёт структурированный draft Scenario.

После генерации:

- draft сохраняется;
- methodist может его редактировать;
- Scenario проходит validation;
- только человек принимает решение о публикации.

AI Scenario Builder не должен автоматически публиковать Scenario.

## 33. Arena

`Arena` — переговоры двух реальных пользователей по опубликованному Scenario.

Arena не является virtual-partner Training.

В Arena:

- два реальных участника;
- у каждого своя роль Scenario;
- общение идёт в реальном времени;
- в текущем MVP используется текстовый chat;
- весь диалог сохраняется;
- после завершения curator/expert может изучить историю;
- feedback даётся отдельно каждому участнику.

Arena создаётся не самим learner, а назначается curator.

## 34. ArenaAssignment

`ArenaAssignment` — запланированное curator событие Arena.

ArenaAssignment связывает:

- curator;
- Scenario;
- конкретную ScenarioVersion;
- participant A;
- participant B;
- роль каждого участника внутри Scenario;
- `scheduled_at`;
- статус;
- связанную ArenaSession после её появления.

Curator выбирает двух пользователей и назначает дату/время.

Пользователь должен видеть свои будущие Arena в профиле/расписании.

Основные состояния ArenaAssignment:

- `scheduled`;
- `ready`;
- `in_progress`;
- `completed`;
- `cancelled`.

Точное временное окно, в котором Arena переходит из `scheduled` в доступное для подключения состояние, может быть определено позднее.

## 35. ArenaSession

`ArenaSession` — фактически запущенная сессия переговоров двух назначенных пользователей.

Не путать с ArenaAssignment:

- ArenaAssignment = план;
- ArenaSession = реально происходящие переговоры.

ArenaSession содержит:

- идентификатор;
- ArenaAssignment;
- ScenarioVersion;
- двух участников;
- channel;
- status;
- время старта;
- время завершения;
- историю сообщений.

Возможные состояния:

- `waiting_for_participants`;
- `in_progress`;
- `completed`;
- `reviewed`.

Первый участник может подключиться раньше второго и ждать.

Когда оба назначенных участника подключились и Arena разрешена по времени, Session может перейти в `in_progress`.

## 36. Arena WebSocket chat

Live-общение Arena идёт через WebSocket.

REST используется для:

- создания ArenaAssignment;
- получения расписания;
- join;
- получения состояния Session;
- получения истории сообщений;
- завершения;
- просмотра и записи feedback.

WebSocket используется для:

- отправки live messages;
- получения новых messages;
- событий подключения/отключения участников;
- событий старта/завершения Arena.

### message_id

Frontend сам генерирует постоянный `message_id` и отправляет:

```json
{
  "type": "message.send",
  "message_id": "uuid-generated-by-client",
  "text": "Что для вас кроме цены наиболее важно?"
}
```

Backend:

1. проверяет idempotency;
2. сохраняет сообщение;
3. рассылает событие `message.created`.

Пример серверного события:

```json
{
  "type": "message.created",
  "message": {
    "message_id": "uuid-generated-by-client",
    "sender_id": "user-uuid",
    "text": "Что для вас кроме цены наиболее важно?",
    "created_at": "..."
  }
}
```

`message_id` одновременно является постоянным идентификатором и ключом идемпотентности. Повторная отправка того же `message_id` не создаёт дубль.

## 37. Arena messages

Каждое сообщение Arena сохраняется.

ArenaMessage содержит как минимум:

- `message_id`;
- `arena_session_id`;
- `sender_id`;
- text;
- created_at.

История сообщений должна быть доступна через обычный REST API даже после завершения WebSocket-соединения.

Это позволяет:

- восстановить чат после reconnect;
- показать историю Arena;
- дать expert возможность провести review;
- в будущем связать transcript с audio.

## 38. Arena history

Пользователь должен иметь два разных представления Arena.

### Запланированные Arena

Показывают будущие и активные Arena:

- Scenario;
- дату/время;
- статус;
- возможность подключиться.

### История Arena

Показывает завершённые Arena:

- Scenario;
- дату;
- второго участника, если это разрешено продуктовой моделью;
- status;
- сохранённые сообщения;
- наличие feedback;
- сам feedback.

В будущем к деталям завершённой Arena могут добавляться media, например аудиозапись.

Текущий chat-only MVP не требует наличия audio.

## 39. Arena feedback

После завершения Arena curator/expert может изучить сохранённый диалог и оставить feedback.

Feedback относится:

- к одной ArenaSession;
- к одному конкретному participant;
- к одному expert/curator.

Минимальная модель feedback:

- текстовый comment;
- created_at.

В будущем feedback может содержать ratings.

Пример расширяемой модели:

```json
{
  "participant_id": "uuid",
  "comment": "Хорошо удерживал границу, но слишком быстро перешёл к уступке.",
  "ratings": [
    {
      "criterion": "Торг и обмен",
      "score": 3,
      "comment": "..."
    }
  ]
}
```

`ratings` пока должны оставаться optional, потому что методика экспертных оценок не зафиксирована.

До определения этой методики экспертные оценки Arena не должны автоматически изменять Profile.

## 40. Будущий voice

В дальнейшем Training и Arena могут получить voice.

Для Arena это означает:

- real-time audio вместо или вместе с text chat;
- сохранение audio/media;
- возможность прослушать запись из Arena history;
- возможную transcript generation.

Текущая предметная модель должна быть расширяемой в эту сторону, но voice не является обязательным для chat-only MVP.

## 41. Важные различия сущностей

### Продуктовая роль vs роль в Scenario

Продуктовые роли:

- `learner`;
- `methodist`;
- `curator`.

Роли сторон переговоров:

- `learner`;
- `partner`.

Это разные понятия.

### Auth session vs TrainingSession vs ArenaSession

Auth session определяет:

- кто вошёл;
- какая продуктовая роль активна.

TrainingSession — индивидуальное прохождение Training.

ArenaSession — переговоры двух реальных пользователей.

Это разные сущности.

### Training Assignment vs ArenaAssignment

Training Assignment — назначение curator одной Training одному learner.

ArenaAssignment — запланированные переговоры двух реальных пользователей в конкретную дату/время.

### ArenaAssignment vs ArenaSession

ArenaAssignment = запланированное событие.

ArenaSession = фактическая история переговоров после подключения участников.

### Scenario vs Session

Scenario = описание переговорной ситуации и правил.

TrainingSession / ArenaSession = конкретная история использования ScenarioVersion.

### Scenario vs ScenarioVersion

Scenario = редактируемая сущность.

ScenarioVersion = immutable опубликованный snapshot.

### Outcome vs Competency

Outcome = результат сделки.

Competency = наблюдаемое качество переговорного поведения.

Одно не должно подменять другое.

### Virtual partner vs Arena partner

Virtual partner — AI-партнёр внутри `virtual_partner` Training.

Arena partner — другой реальный пользователь.

Не смешивать эти сущности.

## 42. Текущие границы MVP

В текущем рабочем контуре приоритетны:

- регистрация и авторизация;
- организации;
- роли и переключение active role;
- создание Scenario;
- AI generation Scenario;
- редактирование и validation;
- публикация и versioning;
- deterministic Training;
- virtual-partner Training в chat;
- сохранение Training history;
- Training Analysis;
- Profile;
- Recommendation;
- возможности curator: команда, Profile/history, Training Assignment, аналитика;
- планирование Arena curator;
- два реальных участника Arena;
- расписание Arena у пользователя;
- join назначенной Arena;
- WebSocket text chat;
- сохранение Arena messages;
- история Arena;
- экспертный текстовый feedback.

Не следует автоматически считать частью текущего MVP:

- voice calls;
- аудиозаписи;
- video/meeting;
- отдельную роль expert;
- жёстко определённую систему Arena ratings;
- автоматическое обновление Profile из Arena;
- intention/perception workflow;
- сложную геймификацию;
- RPG skill tree.

## 43. Инварианты предметной модели

При дальнейших изменениях следует сохранять следующие принципы:

1. Не смешивать продуктовые роли и роли внутри переговоров.
2. Не смешивать auth session, TrainingSession и ArenaSession.
3. Не смешивать Training Assignment и ArenaAssignment.
4. ArenaAssignment является планом, ArenaSession — фактической историей переговоров.
5. Старые TrainingSession и ArenaSession должны ссылаться на опубликованную ScenarioVersion.
6. Изменение Scenario не должно менять историю завершённых Training/Arena.
7. Хороший outcome не является доказательством высокой competency.
8. В deterministic Training scoring должен быть детерминированным и опираться на observation rules.
9. В virtual-partner Training LLM может классифицировать свободный текст и генерировать ответ партнёра, но не должна быть непрозрачным единственным источником competency score.
10. Arena всегда предполагает двух реальных участников.
11. Arena messages должны сохраняться независимо от WebSocket-соединения.
12. Frontend генерирует `message_id`; backend сохраняет его как постоянный идентификатор и ключ идемпотентности.
13. Экспертный feedback относится к конкретному участнику конкретной ArenaSession.
14. Arena ratings пока optional и не должны менять Profile автоматически.
15. AI Scenario Builder предлагает draft, но не публикует Scenario без участия methodist.
16. Новые поля и механики не должны ретроспективно менять историю уже завершённых Training/Arena.
