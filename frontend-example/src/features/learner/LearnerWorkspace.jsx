import { useEffect, useMemo, useRef, useState } from "react";
import { api } from "../../api/client.js";
import { EmptyState, ErrorMessage, Loading } from "../../components/Feedback.jsx";
import { useResource } from "../../hooks/useResource.js";

const PROFILE_MOCK = [
  ["Цели и границы", 3, "Удерживает целевой коридор"],
  ["Диагностика", 2, "Уточняет интересы партнёра"],
  ["Создание ценности", 2, "Ищет обмен по нескольким параметрам"],
  ["Аргументация", 3, "Связывает аргумент с интересом"],
  ["Торг и обмен", 2, "Запрашивает встречную ценность"],
  ["Конфликт и регуляция", 3, "Сохраняет рабочий тон под давлением"],
];

export function LearnerWorkspace({ user }) {
  const [tab, setTab] = useState("overview");
  const [training, setTraining] = useState(null);
  const [arena, setArena] = useState(null);
  const [error, setError] = useState("");
  const assignments = useResource(() => api.arenaAssignments(user.id), [user.id]);
  const arenaHistory = useResource(() => api.arenaHistory(user.id), [user.id]);
  const scenarios = useResource(api.scenarios, []);
  const profileRange = useMemo(() => dateRange(180), [user.id]);
  const profile = useResource(() => api.profileHistory(user.id, profileRange), [user.id, profileRange.from, profileRange.to]);

  const startTraining = async (scenario, mode) => {
    setError("");
    try {
      const session = await api.startTraining({ scenario_id: scenario.scenario_id, mode, channel: "chat" });
      setTraining({ ...session, scenario_title: scenario.title });
    } catch (err) { setError(err.message); }
  };

  const joinArena = async (assignmentID) => {
    setError("");
    try { setArena(await api.joinArena(assignmentID)); } catch (err) { setError(err.message); }
  };

  if (arena) return <ArenaRoom session={arena} userID={user.id} onClose={async () => { setArena(null); await Promise.all([assignments.refresh(), arenaHistory.refresh()]); }} />;
  if (training) return <TrainingRoom session={training} onClose={() => setTraining(null)} />;

  const playable = scenarios.data?.scenarios?.filter((item) => item.status === "published") || [];
  const skills = profile.data?.clusters?.map((item) => [item.cluster, item.after?.level ?? 0, item.after?.indicator]) || PROFILE_MOCK;

  return (
    <>
      <header className="page-head split-head">
        <div><p className="eyebrow">Ученик</p><h1>Моя практика</h1><p>Профиль навыков, Training, расписание Arena, история и все экспертные feedback.</p></div>
        <div className="head-stats"><span><strong>{assignments.data?.arenas?.length || 0}</strong><small>Arena впереди</small></span><span><strong>{arenaHistory.data?.arenas?.length || 0}</strong><small>Arena завершено</small></span></div>
      </header>
      <ErrorMessage>{error}</ErrorMessage>
      <div className="tabs page-tabs">
        {[["overview","Обзор"],["training","Тренировки"],["arena","Arena"],["history","История"],["feedback","Feedback"]].map(([id,label]) => <button key={id} className={tab===id?"active":""} onClick={() => setTab(id)}>{label}</button>)}
      </div>

      {tab === "overview" && <LearnerOverview user={user} skills={skills} profile={profile} assignments={assignments} history={arenaHistory} />}
      {tab === "training" && <TrainingLibrary scenarios={playable} loading={scenarios.loading} error={scenarios.error} onStart={startTraining} />}
      {tab === "arena" && <ArenaSchedule assignments={assignments} onJoin={joinArena} />}
      {tab === "history" && <LearnerHistory userID={user.id} arenaHistory={arenaHistory} />}
      {tab === "feedback" && <MyFeedback userID={user.id} arenaHistory={arenaHistory} />}
    </>
  );
}

function LearnerOverview({ user, skills, profile, assignments, history }) {
  return (
    <div className="overview-grid">
      <section className="profile-panel">
        <div className="panel-title"><div><p className="eyebrow">Профиль</p><strong>Переговорные навыки</strong></div><span className="avatar large">{user.username.slice(0,2).toUpperCase()}</span></div>
        {profile.loading && <Loading />}
        {!profile.data?.clusters?.length && !profile.loading && <div className="mock-badge"><strong>MVP</strong><span>Отдельной ручки текущего Profile нет; ниже демонстрационные значения. Profile history используется, когда backend возвращает точки.</span></div>}
        <div className="skill-grid">{skills.map(([cluster, level, indicator]) => <article className="skill-card" key={cluster}><div><strong>{cluster}</strong><small>{indicator || "Текущий уровень"}</small></div><div className="skill-score"><span>{level}/4</span><div className="progress"><i style={{width:`${level*25}%`}} /></div></div></article>)}</div>
      </section>
      <aside className="dashboard-side">
        <section className="recommendation-card"><p className="eyebrow">Следующая рекомендация</p><h2>Торг и обмен</h2><p>Потренируйте запрос встречной ценности после уступки.</p><div className="mock-badge"><strong>MVP</strong><span>Recommendation endpoint отсутствует; карточка демонстрационная.</span></div></section>
        <section className="mini-list"><div className="panel-title"><strong>Ближайшая Arena</strong></div>{assignments.data?.arenas?.[0] ? <><strong>{assignments.data.arenas[0].scenario_title}</strong><small>{new Date(assignments.data.arenas[0].scheduled_at).toLocaleString()}</small></> : <small>Пока не назначена</small>}</section>
        <section className="mini-list"><div className="panel-title"><strong>Последняя Arena</strong></div>{history.data?.arenas?.[0] ? <><strong>{history.data.arenas[0].scenario_title}</strong><small>{history.data.arenas[0].has_feedback ? "Feedback готов" : "Feedback ожидается"}</small></> : <small>История пуста</small>}</section>
      </aside>
    </div>
  );
}

function TrainingLibrary({ scenarios, loading, error, onStart }) {
  return (
    <section className="training-library">
      <div className="section-heading"><div><p className="eyebrow">Training</p><h2>Доступные сценарии</h2></div></div>
      <div className="mvp-banner"><strong>Training assignments.</strong> Backend умеет создавать назначение, но не содержит GET списка назначений learner. Поэтому библиотека показывает опубликованные Scenario; отдельный блок «назначено куратором» появится после добавления ручки.</div>
      {loading && <Loading />}{error && <ErrorMessage>{error}</ErrorMessage>}
      <div className="cards-grid">{scenarios.map((scenario) => <article className="practice-card expanded" key={scenario.scenario_id}><div><p className="eyebrow">{scenario.difficulty} · v{scenario.version}</p><h3>{scenario.title}</h3><small>{scenario.channels.join(", ")}</small></div><div className="inline-actions"><button onClick={() => onStart(scenario,"deterministic")}>Детерминированная</button><button className="primary" onClick={() => onStart(scenario,"virtual_partner")}>AI-партнёр</button></div></article>)}</div>
      {!loading && !scenarios.length && <EmptyState title="Нет опубликованных сценариев" detail="После публикации они появятся здесь." />}
    </section>
  );
}

function ArenaSchedule({ assignments, onJoin }) {
  return <section className="arena-schedule"><div className="section-heading"><div><p className="eyebrow">Arena</p><h2>Запланированные переговоры</h2></div><button onClick={assignments.refresh}>↻ Обновить</button></div>{assignments.loading && <Loading />}{assignments.error && <ErrorMessage>{assignments.error}</ErrorMessage>}<div className="timeline-list">{assignments.data?.arenas?.map((item) => <article className="arena-assignment-card" key={item.arena_assignment_id}><div className="date-tile"><strong>{new Date(item.scheduled_at).getDate()}</strong><small>{new Date(item.scheduled_at).toLocaleString("ru",{month:"short"})}</small></div><div><strong>{item.scenario_title}</strong><small>{new Date(item.scheduled_at).toLocaleString()} · {item.status}</small></div><button className="primary" onClick={() => onJoin(item.arena_assignment_id)} disabled={["completed","cancelled"].includes(item.status)}>Подключиться</button></article>)}</div>{!assignments.loading && !assignments.data?.arenas?.length && <EmptyState title="Нет запланированных Arena" detail="Куратор выберет двух участников, сценарий и дату." />}</section>;
}

function LearnerHistory({ userID, arenaHistory }) {
  const [details, setDetails] = useState(null);
  const [error, setError] = useState("");
  const openArena = async (id) => { try { setError(""); setDetails(await api.arenaHistoryDetails(userID,id)); } catch (err) { setError(err.message); } };
  return <div className="history-columns"><section><div className="section-heading"><div><p className="eyebrow">Training history</p><h2>Тренировки</h2></div></div><div className="mock-badge"><strong>MVP</strong><span>В текущем API нет GET истории Training сессий пользователя. Показывается макет будущего списка.</span></div><div className="timeline-list"><article className="history-card"><div><strong>Продление контракта</strong><small>deterministic · завершено</small></div><p>Здесь будет переход к transcript/event log и Training Analysis.</p></article><article className="history-card"><div><strong>Работа с возражением</strong><small>virtual partner · завершено</small></div><p>Здесь будет сохранённый chat и semantic observations.</p></article></div></section><section><div className="section-heading"><div><p className="eyebrow">Arena history</p><h2>Живые переговоры</h2></div></div>{arenaHistory.loading && <Loading />}<ErrorMessage>{arenaHistory.error || error}</ErrorMessage><div className="timeline-list">{arenaHistory.data?.arenas?.map((item) => <button className="history-card clickable" key={item.arena_session_id} onClick={() => openArena(item.arena_session_id)}><div><strong>{item.scenario_title}</strong><small>{item.completed_at ? new Date(item.completed_at).toLocaleString() : item.status}</small></div><span>{item.has_feedback ? "Feedback ✓" : "Без feedback"}</span></button>)}</div>{details && <ArenaHistoryDetails details={details} userID={userID} />}</section></div>;
}

function MyFeedback({ userID, arenaHistory }) {
  const [feedback, setFeedback] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  useEffect(() => {
    let alive = true;
    const rows = (arenaHistory.data?.arenas || []).filter((x) => x.has_feedback);
    if (!rows.length) { setFeedback([]); setLoading(false); return; }
    setLoading(true);
    Promise.all(rows.map((row) => api.arenaHistoryDetails(userID,row.arena_session_id))).then((details) => { if (alive) setFeedback(details.flatMap((detail) => detail.feedback.map((item) => ({...item, scenario_title:detail.session.scenario_title}))).filter((item) => item.participant_id === userID)); }).catch((err) => alive && setError(err.message)).finally(() => alive && setLoading(false));
    return () => { alive = false; };
  }, [arenaHistory.data,userID]);
  return <section className="feedback-feed"><div className="section-heading"><div><p className="eyebrow">Экспертный feedback</p><h2>Все отзывы по Arena</h2></div></div>{loading && <Loading />}<ErrorMessage>{error}</ErrorMessage>{feedback.map((item) => <article className="feedback-card" key={item.feedback_id}><div><strong>{item.scenario_title}</strong><small>{new Date(item.created_at).toLocaleString()}</small></div><p>{item.comment}</p>{item.ratings?.length > 0 && <div className="rating-row">{item.ratings.map((r) => <span key={r.criterion}>{r.criterion}: <b>{r.score}</b></span>)}</div>}</article>)}{!loading && !feedback.length && <EmptyState title="Feedback пока нет" detail="После review завершённой Arena комментарии к вашей игре появятся здесь." />}</section>;
}

function TrainingRoom({ session, onClose }) {
  const [current, setCurrent] = useState(session);
  const [analysis, setAnalysis] = useState(null);
  const [text, setText] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const selectAction = async (action) => {
    setBusy(true); setError("");
    try {
      const result = await api.trainingAction(current.session_id, { action_id: action.action_id, client_event_id: crypto.randomUUID() });
      setCurrent((prev) => ({ ...prev, current_node: result.next_node, status: result.status }));
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  };

  const sendMessage = async () => {
    const body = text.trim(); if (!body) return;
    const message_id = crypto.randomUUID();
    setBusy(true); setError("");
    try {
      const result = await api.trainingMessage(current.session_id, { text: body, message_id });
      setCurrent((prev) => ({ ...prev, messages: upsertMessages(prev.messages || [], [result.learner_message, result.partner_message]) }));
      setText("");
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  };

  const complete = async () => {
    setBusy(true); setError("");
    try { await api.completeTraining(current.session_id); setAnalysis(await api.trainingAnalysis(current.session_id)); } catch (err) { setError(err.message); } finally { setBusy(false); }
  };

  if (analysis) return <section className="training-room"><button onClick={onClose}>← К практике</button><p className="eyebrow">Training Analysis</p><h1>{current.scenario_title || "Тренировка завершена"}</h1><AnalysisView analysis={analysis} /></section>;

  return <section className="training-room"><header className="room-head"><button onClick={onClose}>← К практике</button><div><p className="eyebrow">Training · {current.mode}</p><h2>{current.scenario_title || "Переговорная тренировка"}</h2></div><span className="status-pill draft">{current.status}</span></header>{current.mode === "virtual_partner" ? <div className="virtual-training"><div className="chat-stream">{current.messages?.map((message) => <div className={`chat-bubble ${message.author === "learner" ? "mine" : "partner"}`} key={message.message_id}><small>{message.author === "learner" ? "Вы" : "AI-партнёр"}</small><p>{message.text}</p></div>)}</div><div className="chat-compose"><textarea value={text} onChange={(e) => setText(e.target.value)} placeholder="Напишите реплику партнёру" /><button className="primary" disabled={busy || !text.trim()} onClick={sendMessage}>Отправить</button></div></div> : <div className="deterministic-training"><div className="partner-line"><small>Партнёр</small><h2>{current.current_node?.text || "Сценарий готов к завершению"}</h2></div><div className="action-stack">{current.current_node?.available_actions?.map((action) => <button className="action-option" disabled={busy} key={action.action_id} onClick={() => selectAction(action)}><span>{action.label}</span><small>{action.semantic_action || "выбрать"}</small></button>)}</div></div>}<ErrorMessage>{error}</ErrorMessage><div className="room-footer"><button className="primary" disabled={busy} onClick={complete}>Завершить и получить разбор</button></div></section>;
}

function AnalysisView({ analysis }) {
  return <div className="analysis-view"><section><p className="eyebrow">Result</p><pre className="json-box">{JSON.stringify(analysis.result,null,2)}</pre></section><section><p className="eyebrow">Process</p><div className="timeline-list">{analysis.process?.map((step) => <article className="history-card" key={step.step}><div><strong>Шаг {step.step}</strong><small>{step.semantic_actions.join(", ")}</small></div><p>{step.summary}</p></article>)}</div></section><section><p className="eyebrow">Competency</p><div className="skill-grid">{analysis.competency?.map((item,index) => <article className="skill-card" key={`${item.cluster}-${item.indicator}-${index}`}><div><strong>{item.cluster}</strong><small>{item.indicator}</small></div><div className="skill-score"><span>{item.level}/4</span><div className="progress"><i style={{width:`${item.level*25}%`}} /></div></div><p>{item.note}</p></article>)}</div></section></div>;
}

function ArenaRoom({ session, userID, onClose }) {
  const messages = useResource(() => api.arenaMessages(session.arena_session_id), [session.arena_session_id]);
  const socketRef = useRef(null);
  const [text, setText] = useState("");
  const [socketError, setSocketError] = useState("");
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const token = api.token();
    const socket = new WebSocket(api.arenaWebSocketURL(session.arena_session_id, session.websocket_path), token ? ["arena-v1", token] : ["arena-v1"]);
    socketRef.current = socket;
    socket.onopen = () => { setConnected(true); setSocketError(""); };
    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === "message.created") messages.setData((data) => ({ messages: upsertMessages(data?.messages || [], [payload.message]) }));
        if (payload.type === "arena.completed") setSocketError("Arena завершена.");
        if (payload.type === "error") setSocketError(payload.message);
      } catch { setSocketError("Получено некорректное событие WebSocket."); }
    };
    socket.onerror = () => setSocketError("Не удалось подключиться к live-чату.");
    socket.onclose = () => setConnected(false);
    return () => socket.close();
  }, [session.arena_session_id, session.websocket_path, messages.setData]);

  const send = () => {
    const body = text.trim();
    if (!body || socketRef.current?.readyState !== WebSocket.OPEN) return;
    socketRef.current.send(JSON.stringify({ type: "message.send", message_id: crypto.randomUUID(), text: body }));
    setText("");
  };

  return <section className="arena-room"><header className="room-head"><button onClick={onClose}>← К практике</button><div><p className="eyebrow">Arena · {connected ? "live" : session.status}</p><h2>Живые переговоры</h2></div><span className={connected ? "live-badge" : "status-pill draft"}>{connected ? "● LIVE" : session.status}</span></header><div className="chat-stream arena-chat">{messages.data?.messages?.map((message) => <div className={`chat-bubble ${message.sender_id === userID ? "mine" : "partner"}`} key={message.message_id}><small>{message.sender_id === userID ? "Вы" : "Партнёр"}</small><p>{message.text}</p><time>{new Date(message.created_at).toLocaleTimeString()}</time></div>)}</div><ErrorMessage>{messages.error || socketError}</ErrorMessage><div className="chat-compose"><textarea value={text} onChange={(e) => setText(e.target.value)} onKeyDown={(e) => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); send(); } }} placeholder="Введите сообщение" /><button className="primary" disabled={!connected || !text.trim()} onClick={send}>Отправить</button></div><small className="idempotency-note">Каждое сообщение получает frontend UUID в поле <code>message_id</code>; повторная отправка того же ID не должна создавать дубль.</small></section>;
}

function ArenaHistoryDetails({ details, userID }) {
  const myFeedback = details.feedback.filter((item) => item.participant_id === userID);
  return <div className="arena-details"><div className="panel-title"><strong>{details.session.scenario_title} · v{details.session.scenario_version}</strong><span>{details.session.status}</span></div><div className="chat-stream compact-stream">{details.messages.map((m) => <div className={`chat-line ${m.sender_id===userID?"mine-line":""}`} key={m.message_id}><small>{m.sender_id===userID?"Вы":"Партнёр"}</small><p>{m.text}</p></div>)}</div><h4>Ваш feedback</h4>{myFeedback.length ? myFeedback.map((f) => <article className="feedback-item" key={f.feedback_id}><p>{f.comment}</p><small>{new Date(f.created_at).toLocaleString()}</small></article>) : <small>Feedback пока не оставлен.</small>}{details.media?.length ? <div className="media-list">{details.media.map((media) => <a key={media.media_id} href={media.url} target="_blank" rel="noreferrer">Аудиозапись · {media.duration_seconds || 0} сек.</a>)}</div> : <div className="mock-badge"><strong>MVP</strong><span>Audio/media отсутствует: текущий обязательный канал — chat.</span></div>}</div>;
}

function upsertMessages(messages, incoming) {
  const map = new Map(messages.map((item) => [item.message_id, item]));
  incoming.filter(Boolean).forEach((item) => map.set(item.message_id, item));
  return [...map.values()].sort((a,b) => new Date(a.created_at || 0) - new Date(b.created_at || 0));
}
function dateRange(days) { const to = new Date(); const from = new Date(); from.setDate(from.getDate()-days); return { from:from.toISOString().slice(0,10), to:to.toISOString().slice(0,10) }; }
