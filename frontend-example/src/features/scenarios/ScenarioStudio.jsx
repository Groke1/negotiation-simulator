import { useEffect, useMemo, useState } from "react";
import { api } from "../../api/client.js";
import { EmptyState, ErrorMessage, Loading } from "../../components/Feedback.jsx";
import { useResource } from "../../hooks/useResource.js";

const CLUSTERS = [
  "Цели и границы",
  "Диагностика",
  "Создание ценности",
  "Аргументация",
  "Торг и обмен",
  "Конфликт и регуляция",
];
const DEFAULT_PARTNER_BEHAVIOR = { persona: "Деловой партнёр", communication_style: "спокойный, предметный", initial_position: "Отстаивает свои исходные условия", pressure_policy: "Усиливает давление при отсутствии аргументов", concession_policy: "Уступает только при встречной ценности", forbidden_behavior: "Не раскрывает private information без основания" };

const SEMANTIC_ACTIONS = [
  "ASK_QUESTION",
  "MAKE_PROPOSAL",
  "MAKE_CONCESSION",
  "REQUEST_RECIPROCITY",
  "HOLD_BOUNDARY",
  "CHANGE_TOPIC",
  "PAUSE",
  "SUMMARIZE",
  "CLOSE_NEGOTIATION",
];

function freshScenario() {
  return {
    version: "0.1",
    title: "Новая переговорная тренировка",
    mode_support: ["training", "arena"],
    context: { situation_type: "переговоры", conflict_type: ["условия"], description: "Опишите исходную ситуацию." },
    roles: [
      { id: "learner", label: "Менеджер" },
      { id: "partner", label: "Партнёр" },
    ],
    goals: { learner: "Цель участника", partner: "Цель партнёра" },
    interests: { learner: ["Сохранить ценность"], partner: ["Получить лучшие условия"] },
    constraints: { learner: { notes: ["Укажите ограничения"] }, partner: { notes: ["Укажите ограничения"] } },
    parameters: [{ key: "условие", initial: "базовое", unit: "" }],
    private_information: { learner: { notes: [] }, partner: { notes: [] } },
    partner_behavior: structuredClone(DEFAULT_PARTNER_BEHAVIOR),
    difficulty: "L1",
    competencies: ["Диагностика", "Цели и границы"],
    observation_rules: [
      { semantic_action: "ASK_QUESTION", cluster: "Диагностика", indicator: "Уточняет интересы", trigger_condition: "Задан содержательный уточняющий вопрос" },
    ],
    completion_conditions: [{ type: "agreement", condition: "Стороны зафиксировали итог переговоров" }],
    channels: ["chat"],
    nodes: [
      {
        id: "n1",
        type: "opening",
        speaker: "partner",
        text: "Начальная реплика партнёра.",
        emotion: "нейтрально",
        actions: [
          { id: "a1", label: "Уточнить интересы партнёра", semantic_action: "ASK_QUESTION", next_node_id: "n2", indicator_tags: [] },
        ],
        is_end: false,
        outcome: null,
      },
      {
        id: "n2",
        type: "resolution",
        speaker: "partner",
        text: "Партнёр готов зафиксировать договорённость.",
        actions: [{ id: "a2", label: "Резюмировать договорённости", semantic_action: "SUMMARIZE", next_node_id: null, indicator_tags: [] }],
        is_end: true,
        outcome: { summary: "Переговоры завершены" },
      },
    ],
  };
}

function cleanScenario(value) {
  const copy = structuredClone(value);
  delete copy.scenario_id;
  delete copy.status;
  delete copy.updated_at;
  // partner_behavior описан в AGENT.md, но отсутствует в текущем scenarios.yaml.
  delete copy.partner_behavior;
  return copy;
}

export function ScenarioStudio({ role = "methodist", embedded = false }) {
  const library = useResource(api.scenarios, []);
  const [selected, setSelected] = useState(null);
  const [draft, setDraft] = useState(null);
  const [versions, setVersions] = useState([]);
  const [tab, setTab] = useState("basics");
  const [prompt, setPrompt] = useState("");
  const [validation, setValidation] = useState(null);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [snapshot, setSnapshot] = useState(null);

  const open = async (item) => {
    setSelected(item);
    setError("");
    setNotice("");
    setValidation(null);
    setSnapshot(null);
    try {
      const [scenario, versionData] = await Promise.all([
        api.scenario(item.scenario_id),
        api.scenarioVersions(item.scenario_id),
      ]);
      setDraft({ ...scenario, status: item.status, partner_behavior: scenario.partner_behavior || structuredClone(DEFAULT_PARTNER_BEHAVIOR) });
      setVersions(versionData.versions || []);
      setTab("basics");
    } catch (err) {
      setError(err.message);
    }
  };

  const manual = () => {
    setSelected(null);
    setVersions([]);
    setValidation(null);
    setSnapshot(null);
    setDraft(freshScenario());
    setTab("basics");
    setNotice("Создан локальный черновик. Нажмите «Сохранить черновик», чтобы записать его на backend.");
  };

  const generate = async (event) => {
    event.preventDefault();
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const result = await api.generateScenario(prompt.trim());
      setPrompt("");
      setNotice("AI создал и сохранил draft. Проверьте поля, граф и правила перед публикацией.");
      await library.refresh();
      const generated = result.scenario_draft;
      const summary = { ...generated, status: "draft" };
      setSelected(summary);
      setDraft({ ...summary, partner_behavior: summary.partner_behavior || structuredClone(DEFAULT_PARTNER_BEHAVIOR) });
      const versionData = await api.scenarioVersions(generated.scenario_id).catch(() => ({ versions: [] }));
      setVersions(versionData.versions || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const save = async () => {
    if (!draft) return;
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const payload = cleanScenario(draft);
      if (draft.scenario_id) {
        await api.updateScenario(draft.scenario_id, payload);
        setNotice("Промежуточный сценарий сохранён. Предыдущие опубликованные snapshot-версии не изменены.");
      } else {
        const result = await api.createScenario(payload);
        setDraft((current) => ({ ...current, scenario_id: result.id, status: "draft" }));
        setSelected({ scenario_id: result.id, title: draft.title, version: draft.version, status: "draft", difficulty: draft.difficulty, channels: draft.channels });
        setNotice("Черновик создан и сохранён.");
      }
      await library.refresh();
    } catch (err) {
      if (role === "curator" && err.status === 403) {
        setError("Backend разрешает запись Scenario только active role=methodist. Конструктор на фронте готов, но для сохранения/AI-генерации куратору нужна backend-permission или роль methodist.");
      } else {
        setError(err.message);
      }
    } finally {
      setBusy(false);
    }
  };

  const validate = async () => {
    if (!draft?.scenario_id) {
      setError("Сначала сохраните черновик, затем запускайте backend validation.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const result = await api.validateScenario(draft.scenario_id);
      setValidation(result);
      setNotice(result.valid ? "Backend validation пройден." : "Validation завершена: есть замечания к структуре/графу/покрытию.");
      setTab("validation");
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const publish = async () => {
    if (!draft?.scenario_id) return;
    setBusy(true);
    setError("");
    try {
      await api.publishScenario(draft.scenario_id);
      setNotice(`Опубликована immutable-версия ${draft.version}.`);
      await library.refresh();
      const versionData = await api.scenarioVersions(draft.scenario_id);
      setVersions(versionData.versions || []);
      setSelected((current) => current ? { ...current, status: "published", version: draft.version } : current);
      setDraft((current) => ({ ...current, status: "published" }));
      setTab("versions");
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const archive = async () => {
    if (!draft?.scenario_id) return;
    setBusy(true);
    setError("");
    try {
      await api.archiveScenario(draft.scenario_id);
      setNotice("Сценарий архивирован; опубликованные версии остаются в истории.");
      setDraft((current) => ({ ...current, status: "archived" }));
      await library.refresh();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };

  const loadSnapshot = async (version) => {
    if (!draft?.scenario_id) return;
    setError("");
    try {
      setSnapshot(await api.scenarioVersion(draft.scenario_id, version));
    } catch (err) {
      setError(err.message);
    }
  };

  const status = selected?.status || draft?.status || "draft";

  return (
    <section className={embedded ? "scenario-studio embedded" : "scenario-studio"}>
      <header className="studio-head">
        <div>
          <p className="eyebrow">Конструктор сценариев</p>
          <h1>{embedded ? "Сценарии" : "Конструктор тренировок"}</h1>
          <p>Ручной draft, AI generation, промежуточное сохранение, validation, публикация и immutable history.</p>
        </div>
        <button className="primary" onClick={manual}>+ Новый вручную</button>
      </header>

      {role === "curator" && (
        <div className="mvp-banner">
          <strong>Ограничение backend-роли.</strong> API сценариев разрешает создание/изменение только active role=methodist. Редактор доступен куратору, но сохранение потребует расширения permission на backend.
        </div>
      )}

      <div className="studio-grid">
        <aside className="library-panel">
          <div className="panel-title"><strong>Библиотека</strong><button onClick={library.refresh}>↻</button></div>
          {library.loading && <Loading />}
          {library.error && <ErrorMessage>{library.error}</ErrorMessage>}
          {library.data?.scenarios?.map((item) => (
            <button className={`library-item ${draft?.scenario_id === item.scenario_id ? "active" : ""}`} key={item.scenario_id} onClick={() => open(item)}>
              <span><strong>{item.title}</strong><small>v{item.version} · {item.difficulty}</small></span>
              <em className={`status-pill ${item.status}`}>{item.status}</em>
            </button>
          ))}
          {!library.loading && !library.data?.scenarios?.length && <EmptyState title="Библиотека пуста" detail="Создайте сценарий вручную или через AI." />}
        </aside>

        <div className="studio-main">
          <form className="ai-builder" onSubmit={generate}>
            <div><p className="eyebrow">AI Scenario Builder</p><strong>Сгенерировать структурированный draft</strong></div>
            <textarea minLength="20" required value={prompt} onChange={(e) => setPrompt(e.target.value)} placeholder="Например: продление B2B-контракта, клиент требует скидку 15%, менеджер может дать максимум 7%..." />
            <button className="primary" disabled={busy}>{busy ? "Генерируем…" : "Сгенерировать нейронкой"}</button>
          </form>

          <ErrorMessage>{error}</ErrorMessage>
          {notice && <p className="notice">{notice}</p>}

          {!draft ? (
            <EmptyState title="Выберите сценарий" detail="Откройте существующий сценарий, создайте новый вручную или сгенерируйте draft." />
          ) : (
            <>
              <div className="editor-toolbar">
                <div><span className={`status-pill ${status}`}>{status}</span><strong>{draft.title}</strong><small> текущая версия: {draft.version}</small></div>
                <div className="inline-actions">
                  <button onClick={save} disabled={busy}>Сохранить черновик</button>
                  <button onClick={validate} disabled={busy}>Проверить</button>
                  <button className="primary" onClick={publish} disabled={busy || !draft.scenario_id}>Опубликовать</button>
                  {draft.scenario_id && <button className="danger" onClick={archive} disabled={busy}>Архивировать</button>}
                </div>
              </div>
              {status === "published" && <div className="version-note">Изменения редактируемого Scenario не меняют уже опубликованные версии. Перед следующей публикацией задайте новую версию.</div>}
              <div className="tabs compact-tabs">
                {[["basics","Основное"],["graph","Граф"],["rules","Правила"],["validation","Validation"],["versions","Версии"]].map(([id,label]) => <button key={id} className={tab===id?"active":""} onClick={() => setTab(id)}>{label}</button>)}
              </div>
              {tab === "basics" && <BasicsEditor draft={draft} setDraft={setDraft} />}
              {tab === "graph" && <GraphEditor draft={draft} setDraft={setDraft} />}
              {tab === "rules" && <RulesEditor draft={draft} setDraft={setDraft} />}
              {tab === "validation" && <ValidationPanel value={validation} />}
              {tab === "versions" && <VersionsPanel versions={versions} snapshot={snapshot} onLoad={loadSnapshot} />}
            </>
          )}
        </div>
      </div>
    </section>
  );
}

function BasicsEditor({ draft, setDraft }) {
  const roleMap = useMemo(() => Object.fromEntries((draft.roles || []).map((role) => [role.id, role.label])), [draft.roles]);
  const patch = (key, value) => setDraft((current) => ({ ...current, [key]: value }));
  const setRole = (id, label) => patch("roles", [
    { id: "learner", label: id === "learner" ? label : roleMap.learner || "Learner" },
    { id: "partner", label: id === "partner" ? label : roleMap.partner || "Partner" },
  ]);
  const toggle = (key, value) => patch(key, draft[key].includes(value) ? draft[key].filter((x) => x !== value) : [...draft[key], value]);
  return (
    <div className="editor-section">
      <div className="form-grid two">
        <label>Название<input value={draft.title} onChange={(e) => patch("title", e.target.value)} /></label>
        <label>Версия<input value={draft.version} onChange={(e) => patch("version", e.target.value)} placeholder="1.1" /></label>
        <label>Тип ситуации<input value={draft.context?.situation_type || ""} onChange={(e) => patch("context", { ...draft.context, situation_type: e.target.value })} /></label>
        <label>Конфликты (через запятую)<input value={(draft.context?.conflict_type || []).join(", ")} onChange={(e) => patch("context", { ...draft.context, conflict_type: splitList(e.target.value) })} /></label>
      </div>
      <label>Описание ситуации<textarea value={draft.context?.description || ""} onChange={(e) => patch("context", { ...draft.context, description: e.target.value })} /></label>
      <div className="form-grid two">
        <label>Роль learner<input value={roleMap.learner || ""} onChange={(e) => setRole("learner", e.target.value)} /></label>
        <label>Роль partner<input value={roleMap.partner || ""} onChange={(e) => setRole("partner", e.target.value)} /></label>
        <label>Цель learner<textarea value={draft.goals?.learner || ""} onChange={(e) => patch("goals", { ...draft.goals, learner: e.target.value })} /></label>
        <label>Цель partner<textarea value={draft.goals?.partner || ""} onChange={(e) => patch("goals", { ...draft.goals, partner: e.target.value })} /></label>
        <label>Интересы learner<textarea value={(draft.interests?.learner || []).join("\n")} onChange={(e) => patch("interests", { ...draft.interests, learner: splitLines(e.target.value) })} /></label>
        <label>Интересы partner<textarea value={(draft.interests?.partner || []).join("\n")} onChange={(e) => patch("interests", { ...draft.interests, partner: splitLines(e.target.value) })} /></label>
      </div>
      <div className="form-grid three">
        <label>Сложность<select value={draft.difficulty} onChange={(e) => patch("difficulty", e.target.value)}><option>L1</option><option>L2</option><option>L3</option></select></label>
        <CheckGroup title="Режимы" values={["training","arena","live"]} selected={draft.mode_support} onToggle={(value) => toggle("mode_support", value)} />
        <CheckGroup title="Каналы" values={["chat","voice","meeting"]} selected={draft.channels} onToggle={(value) => toggle("channels", value)} />
      </div>
      <CheckGroup title="Компетенции" values={CLUSTERS} selected={draft.competencies || []} onToggle={(value) => toggle("competencies", value)} wide />
      <JsonField title="Ограничения" value={draft.constraints} onChange={(value) => patch("constraints", value)} />
      <JsonField title="Параметры сделки" value={draft.parameters} onChange={(value) => patch("parameters", value)} />
      <JsonField title="Private information" value={draft.private_information} onChange={(value) => patch("private_information", value)} />
    </div>
  );
}

function GraphEditor({ draft, setDraft }) {
  const updateNode = (index, patch) => setDraft((current) => ({ ...current, nodes: current.nodes.map((node, i) => i === index ? { ...node, ...patch } : node) }));
  const addNode = () => setDraft((current) => ({ ...current, nodes: [...current.nodes, { id: `n${current.nodes.length + 1}`, type: "opening", speaker: "partner", text: "Новая реплика", actions: [], is_end: false, outcome: null }] }));
  const removeNode = (index) => setDraft((current) => ({ ...current, nodes: current.nodes.filter((_, i) => i !== index) }));
  const addAction = (nodeIndex) => setDraft((current) => ({ ...current, nodes: current.nodes.map((node, i) => i === nodeIndex ? { ...node, actions: [...node.actions, { id: `a${Date.now().toString(36)}`, label: "Новое действие", semantic_action: "ASK_QUESTION", next_node_id: null, indicator_tags: [] }] } : node) }));
  const updateAction = (nodeIndex, actionIndex, patch) => setDraft((current) => ({ ...current, nodes: current.nodes.map((node, i) => i === nodeIndex ? { ...node, actions: node.actions.map((action, j) => j === actionIndex ? { ...action, ...patch } : action) } : node) }));
  const removeAction = (nodeIndex, actionIndex) => setDraft((current) => ({ ...current, nodes: current.nodes.map((node, i) => i === nodeIndex ? { ...node, actions: node.actions.filter((_, j) => j !== actionIndex) } : node) }));
  return (
    <div className="editor-section">
      <div className="section-heading"><div><h3>Детерминированный граф</h3><p>Узлы, реплики, semantic action и переходы.</p></div><button onClick={addNode}>+ Узел</button></div>
      {(draft.nodes || []).map((node, index) => (
        <article className="node-card" key={`${node.id}-${index}`}>
          <div className="node-head">
            <input value={node.id} onChange={(e) => updateNode(index, { id: e.target.value })} />
            <select value={node.type} onChange={(e) => updateNode(index, { type: e.target.value })}><option value="opening">opening</option><option value="escalation">escalation</option><option value="resolution">resolution</option></select>
            <select value={node.speaker} onChange={(e) => updateNode(index, { speaker: e.target.value })}><option value="partner">partner</option><option value="learner">learner</option></select>
            <label className="inline-check"><input type="checkbox" checked={Boolean(node.is_end)} onChange={(e) => updateNode(index, { is_end: e.target.checked })} /> terminal</label>
            <button className="icon-danger" onClick={() => removeNode(index)}>×</button>
          </div>
          <textarea value={node.text} onChange={(e) => updateNode(index, { text: e.target.value })} />
          <div className="actions-list">
            {node.actions.map((action, actionIndex) => (
              <div className="action-editor" key={`${action.id}-${actionIndex}`}>
                <input className="mono-input" value={action.id} onChange={(e) => updateAction(index, actionIndex, { id: e.target.value })} />
                <input value={action.label} onChange={(e) => updateAction(index, actionIndex, { label: e.target.value })} />
                <select value={action.semantic_action} onChange={(e) => updateAction(index, actionIndex, { semantic_action: e.target.value })}>{SEMANTIC_ACTIONS.map((item) => <option key={item}>{item}</option>)}</select>
                <input placeholder="next node / пусто" value={action.next_node_id || ""} onChange={(e) => updateAction(index, actionIndex, { next_node_id: e.target.value || null })} />
                <button className="icon-danger" onClick={() => removeAction(index, actionIndex)}>×</button>
              </div>
            ))}
            <button onClick={() => addAction(index)}>+ Действие</button>
          </div>
        </article>
      ))}
    </div>
  );
}

function RulesEditor({ draft, setDraft }) {
  const patch = (key, value) => setDraft((current) => ({ ...current, [key]: value }));
  return (
    <div className="editor-section form-grid two">
      <JsonField title="Observation rules" value={draft.observation_rules} onChange={(value) => patch("observation_rules", value)} />
      <JsonField title="Completion conditions" value={draft.completion_conditions} onChange={(value) => patch("completion_conditions", value)} />
      <div className="guide-card">
        <h3>Virtual partner behavior</h3>
        <div className="mock-badge"><strong>MVP</strong><span>Поле partner_behavior есть в предметной модели, но отсутствует в текущем scenarios.yaml. Настройка сохраняется только локально и не отправляется в API.</span></div>
        <JsonField title="Partner behavior (local draft)" value={draft.partner_behavior || DEFAULT_PARTNER_BEHAVIOR} onChange={(value) => patch("partner_behavior", value)} />
      </div>
      <div className="guide-card">
        <h3>Semantic actions</h3>
        <p>{SEMANTIC_ACTIONS.join(" · ")}</p>
      </div>
      <div className="guide-card">
        <h3>Покрытие</h3>
        <p>Validation на backend проверяет структуру, тупики/недостижимые узлы/циклы и заявленное покрытие indicators.</p>
      </div>
    </div>
  );
}

function ValidationPanel({ value }) {
  if (!value) return <EmptyState title="Validation ещё не запускалась" detail="Сохраните draft и нажмите «Проверить»." />;
  const problems = [
    ...(value.graph?.dead_ends || []).map((x) => `Тупик: ${x}`),
    ...(value.graph?.unreachable_nodes || []).map((x) => `Недостижим: ${x}`),
    ...(value.graph?.cycles || []).map((x) => `Цикл: ${x.join(" → ")}`),
  ];
  return (
    <div className="editor-section">
      <div className={`validation-summary ${value.valid ? "ok" : "warn"}`}><strong>{value.valid ? "Сценарий валиден" : "Есть замечания"}</strong><span>valid = {String(value.valid)}</span></div>
      <div className="metric-grid">
        {Object.entries(value.structure || {}).map(([key, flag]) => <div key={key}><small>{key}</small><strong>{flag ? "✓" : "×"}</strong></div>)}
      </div>
      <h3>Граф</h3>
      {problems.length ? problems.map((item) => <p className="validation-line" key={item}>{item}</p>) : <p className="notice">Проблем графа не найдено.</p>}
      <h3>Недостаточное покрытие</h3>
      <pre className="json-box">{JSON.stringify(value.coverage?.under_covered_indicators || [], null, 2)}</pre>
    </div>
  );
}

function VersionsPanel({ versions, snapshot, onLoad }) {
  return (
    <div className="editor-section">
      <h3>Опубликованные immutable snapshots</h3>
      {versions.length ? <div className="version-list">{versions.map((item) => <button key={item.version} onClick={() => onLoad(item.version)}><strong>v{item.version}</strong><small>{new Date(item.published_at).toLocaleString()}</small></button>)}</div> : <EmptyState title="Версий пока нет" detail="Первая версия появится после публикации." />}
      {snapshot && <div><p className="eyebrow">Snapshot v{snapshot.version}</p><pre className="json-box tall">{JSON.stringify(snapshot.snapshot, null, 2)}</pre></div>}
    </div>
  );
}

function JsonField({ title, value, onChange }) {
  const [text, setText] = useState(() => JSON.stringify(value, null, 2));
  const [error, setError] = useState("");
  useEffect(() => setText(JSON.stringify(value, null, 2)), [value]);
  const apply = () => {
    try { onChange(JSON.parse(text)); setError(""); } catch { setError("Некорректный JSON"); }
  };
  return <label className="json-field">{title}<textarea className="mono" value={text} onChange={(e) => setText(e.target.value)} onBlur={apply} />{error && <small className="field-error">{error}</small>}</label>;
}

function CheckGroup({ title, values, selected = [], onToggle, wide = false }) {
  return <div className={wide ? "check-group wide-group" : "check-group"}><small>{title}</small><div>{values.map((value) => <label key={value}><input type="checkbox" checked={selected.includes(value)} onChange={() => onToggle(value)} /><span>{value}</span></label>)}</div></div>;
}

function splitList(value) { return value.split(",").map((x) => x.trim()).filter(Boolean); }
function splitLines(value) { return value.split("\n").map((x) => x.trim()).filter(Boolean); }
