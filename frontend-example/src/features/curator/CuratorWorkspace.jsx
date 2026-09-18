import { useEffect, useMemo, useState } from "react";
import { api } from "../../api/client.js";
import { EmptyState, ErrorMessage, Loading } from "../../components/Feedback.jsx";
import { useResource } from "../../hooks/useResource.js";

const PROFILE_MOCK = [
  ["Цели и границы", 3], ["Диагностика", 2], ["Создание ценности", 2],
  ["Аргументация", 3], ["Торг и обмен", 2], ["Конфликт и регуляция", 3],
];

export function CuratorWorkspace() {
  const [tab, setTab] = useState("team");
  const [selectedUserID, setSelectedUserID] = useState(null);
  const team = useResource(api.curatorTeam, []);
  const organizationUsers = useResource(api.organizationUsers, []);
  const analytics = useResource(api.teamAnalytics, []);
  const scenarios = useResource(api.scenarios, []);
  const arenas = useResource(api.organizationArenas, []);

  const members = useMemo(() => {
    const details = new Map((organizationUsers.data?.users || []).map((u) => [u.id, u]));
    const curatorMembers = team.data?.team || [];
    if (curatorMembers.length) {
      return curatorMembers.map((member) => ({ ...details.get(member.user_id), ...member, id: member.user_id, username: member.name }));
    }
    return (organizationUsers.data?.users || []).filter((u) => u.roles?.includes("learner")).map((u) => ({ ...u, user_id: u.id, name: u.username, weakest_cluster: null, last_session_at: null }));
  }, [team.data, organizationUsers.data]);

  const selected = members.find((member) => member.user_id === selectedUserID) || null;
  const published = scenarios.data?.scenarios?.filter((item) => item.status === "published") || [];

  return (
    <>
      <header className="page-head split-head">
        <div>
          <p className="eyebrow">Куратор</p>
          <h1>Команда, назначения и Arena</h1>
          <p>Профили и динамика сотрудников, Training/Arena assignments, live monitoring и feedback.</p>
        </div>
        <div className="head-stats">
          <span><strong>{members.length}</strong><small>сотрудников</small></span>
          <span><strong>{arenas.data?.arenas?.filter((x) => x.status === "in_progress").length || 0}</strong><small>live Arena</small></span>
        </div>
      </header>

      <div className="tabs page-tabs">
        {[["team","Команда"],["assign","Назначения"],["monitor","Arena monitor"],["analytics","Аналитика"]].map(([id,label]) => (
          <button key={id} className={tab === id ? "active" : ""} onClick={() => setTab(id)}>{label}</button>
        ))}
      </div>

      {tab === "team" && (
        <TeamView members={members} loading={team.loading || organizationUsers.loading} error={team.error || organizationUsers.error} selected={selected} onSelect={(id) => setSelectedUserID(id)} />
      )}
      {tab === "assign" && <AssignmentsView members={members} scenarios={published} onArenaCreated={arenas.refresh} />}
      {tab === "monitor" && <ArenaMonitorList arenas={arenas} members={members} />}
      {tab === "analytics" && <AnalyticsView analytics={analytics} members={members} />}
    </>
  );
}

function TeamView({ members, loading, error, selected, onSelect }) {
  const [filter, setFilter] = useState("");
  const visible = members.filter((member) => `${member.name || member.username} ${member.email || ""} ${member.weakest_cluster || ""}`.toLowerCase().includes(filter.toLowerCase()));
  return (
    <div className="team-workspace">
      <aside className="team-list-panel">
        <div className="panel-title"><strong>Сотрудники</strong><span className="count-badge">{visible.length}</span></div>
        <input value={filter} onChange={(e) => setFilter(e.target.value)} placeholder="Поиск по команде" />
        {loading && <Loading />}
        {error && <ErrorMessage>{error}</ErrorMessage>}
        {!loading && !visible.length && <EmptyState title="Нет сотрудников" detail="Curator team API пока не вернул участников." />}
        <div className="member-list">
          {visible.map((member) => (
            <button key={member.user_id} className={`member-card ${selected?.user_id === member.user_id ? "active" : ""}`} onClick={() => onSelect(member.user_id)}>
              <span className="avatar">{(member.name || member.username || "U").slice(0, 2).toUpperCase()}</span>
              <span className="member-main"><strong>{member.name || member.username}</strong><small>{member.email || "learner"}</small></span>
              <span className="member-meta"><small>слабее всего</small><em>{member.weakest_cluster || "—"}</em></span>
            </button>
          ))}
        </div>
      </aside>
      <section className="member-detail-panel">
        {selected ? <MemberDetail member={selected} /> : <EmptyState title="Откройте профиль сотрудника" detail="Здесь будут навыки, динамика, история Training/Arena и feedback." />}
      </section>
    </div>
  );
}

function MemberDetail({ member }) {
  const [subtab, setSubtab] = useState("profile");
  const range = useMemo(() => dateRange(180), [member.user_id]);
  const profile = useResource(() => api.profileHistory(member.user_id, range), [member.user_id, range.from, range.to]);
  const arenaHistory = useResource(() => api.arenaHistory(member.user_id), [member.user_id]);
  const arenaAssignments = useResource(() => api.arenaAssignments(member.user_id), [member.user_id]);

  const skillRows = profile.data?.clusters?.map((item) => ({ cluster: item.cluster, before: item.before?.level, level: item.after?.level, indicator: item.after?.indicator })) || PROFILE_MOCK.map(([cluster, level]) => ({ cluster, before: Math.max(0, level - 1), level, indicator: "Демонстрационный indicator" }));
  const mockProfile = !profile.data?.clusters?.length;

  return (
    <>
      <div className="member-detail-head">
        <div><p className="eyebrow">Профиль сотрудника</p><h2>{member.name || member.username}</h2><p>{member.email || "Пользователь команды"}</p></div>
        <span className="status-pill published">learner</span>
      </div>
      <div className="tabs compact-tabs">
        {[["profile","Навыки"],["training","Training"],["arena","Arena"],["feedback","Feedback"]].map(([id,label]) => <button key={id} className={subtab === id ? "active" : ""} onClick={() => setSubtab(id)}>{label}</button>)}
      </div>

      {subtab === "profile" && (
        <div className="detail-section">
          {profile.loading && <Loading />}
          {profile.error && <div className="mvp-banner"><strong>Profile history API недоступен для выбранного диапазона.</strong> Ниже показаны демонстрационные значения, чтобы интерфейс профиля был виден в MVP.</div>}
          {mockProfile && !profile.loading && <MockBadge text="Mock: отдельной ручки текущего Profile в API нет" />}
          <div className="skill-grid">{skillRows.map((item) => <SkillCard key={item.cluster} {...item} />)}</div>
          <div className="history-mini"><strong>Последняя тренировка</strong><span>{member.last_session_at ? new Date(member.last_session_at).toLocaleString() : "Нет данных"}</span><strong>Слабый кластер</strong><span>{member.weakest_cluster || "Не определён"}</span></div>
        </div>
      )}

      {subtab === "training" && (
        <div className="detail-section">
          <MockBadge text="Mock: backend не предоставляет GET списка/истории Training пользователя" />
          <div className="timeline-list">
            <HistoryCard title="Переговоры о продлении контракта" meta="12.09 · deterministic · завершено" detail="Разбор: удержание границ стабильно, стоит усилить запрос встречной ценности." />
            <HistoryCard title="Работа с возражением по цене" meta="04.09 · virtual partner · завершено" detail="Разбор и transcript появятся здесь после добавления Training history endpoint." />
          </div>
        </div>
      )}

      {subtab === "arena" && <UserArenaHistory userID={member.user_id} history={arenaHistory} assignments={arenaAssignments} curator />}
      {subtab === "feedback" && <FeedbackOverview userID={member.user_id} history={arenaHistory} />}
    </>
  );
}

function SkillCard({ cluster, before, level, indicator }) {
  const safe = Number.isFinite(level) ? level : 0;
  return <article className="skill-card"><div><strong>{cluster}</strong><small>{indicator || "Текущий уровень"}</small></div><div className="skill-score"><span>{safe}/4</span><div className="progress"><i style={{ width: `${safe * 25}%` }} /></div></div>{Number.isFinite(before) && <small className={safe >= before ? "trend-up" : "trend-down"}>{before} → {safe}</small>}</article>;
}

function AssignmentsView({ members, scenarios, onArenaCreated }) {
  const [mode, setMode] = useState("training");
  const [learner, setLearner] = useState("");
  const [partner, setPartner] = useState("");
  const [scenario, setScenario] = useState("");
  const [scheduledAt, setScheduledAt] = useState("");
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");

  const submit = async (event) => {
    event.preventDefault(); setBusy(true); setNotice(""); setError("");
    try {
      if (mode === "training") {
        const result = await api.createTrainingAssignment({ learner_id: learner, scenario_id: scenario });
        setNotice(`Training назначен: ${result.assignment_id} · ${result.status}`);
      } else {
        const result = await api.createArena({ scenario_id: scenario, scheduled_at: new Date(scheduledAt).toISOString(), participants: [{ user_id: learner, role: "learner" }, { user_id: partner, role: "partner" }] });
        setNotice(`Arena назначена: ${result.arena_assignment_id}`); onArenaCreated();
      }
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  };

  return (
    <section className="assignment-workspace">
      <div className="assignment-card">
        <div className="switcher"><button className={mode === "training" ? "selected" : ""} onClick={() => setMode("training")}>Training</button><button className={mode === "arena" ? "selected" : ""} onClick={() => setMode("arena")}>Arena</button></div>
        <p className="eyebrow">Новое назначение</p><h2>{mode === "training" ? "Назначить тренировку" : "Запланировать Arena"}</h2>
        <form onSubmit={submit}>
          <label>{mode === "training" ? "Сотрудник" : "Участник A (learner)"}<select required value={learner} onChange={(e) => setLearner(e.target.value)}><option value="">Выберите</option>{members.map((m) => <option value={m.user_id} key={m.user_id}>{m.name || m.username}</option>)}</select></label>
          {mode === "arena" && <label>Участник B (partner)<select required value={partner} onChange={(e) => setPartner(e.target.value)}><option value="">Выберите</option>{members.filter((m) => m.user_id !== learner).map((m) => <option value={m.user_id} key={m.user_id}>{m.name || m.username}</option>)}</select></label>}
          <label>Опубликованный сценарий<select required value={scenario} onChange={(e) => setScenario(e.target.value)}><option value="">Выберите</option>{scenarios.map((s) => <option value={s.scenario_id} key={s.scenario_id}>{s.title} · v{s.version}</option>)}</select></label>
          {mode === "arena" && <label>Дата и время<input required type="datetime-local" value={scheduledAt} onChange={(e) => setScheduledAt(e.target.value)} /></label>}
          <ErrorMessage>{error}</ErrorMessage>{notice && <p className="notice">{notice}</p>}
          <button className="primary" disabled={busy || !scenarios.length}>{busy ? "Сохраняем…" : mode === "training" ? "Назначить Training" : "Назначить Arena"}</button>
        </form>
      </div>
      <div className="side-explainer">
        <h3>Два разных типа назначения</h3>
        <p><strong>Training Assignment</strong> — один learner и опубликованный Scenario.</p>
        <p><strong>Arena Assignment</strong> — два реальных участника, ScenarioVersion и scheduled_at.</p>
        <MockBadge text="Список Training assignments отсутствует в API — после POST здесь пока нечего перечитывать с backend" />
      </div>
    </section>
  );
}

function ArenaMonitorList({ arenas, members }) {
  const [selectedID, setSelectedID] = useState(null);
  const rows = arenas.data?.arenas || [];
  const active = rows.find((item) => item.arena_assignment_id === selectedID) || null;
  return (
    <div className="monitor-layout">
      <aside className="monitor-list">
        <div className="panel-title"><strong>Arena организации</strong><button onClick={arenas.refresh}>↻</button></div>
        {arenas.loading && <Loading />}{arenas.error && <ErrorMessage>{arenas.error}</ErrorMessage>}
        {rows.map((arena) => <button key={arena.arena_assignment_id} className={selectedID === arena.arena_assignment_id ? "active" : ""} onClick={() => setSelectedID(arena.arena_assignment_id)}><span><strong>{arena.scenario_title}</strong><small>{new Date(arena.scheduled_at).toLocaleString()}</small></span><em className={`status-dot ${arena.status}`}>{arena.status}</em></button>)}
        {!rows.length && <EmptyState title="Arena нет" detail="Назначьте Arena, чтобы она появилась в monitor." />}
      </aside>
      <section className="monitor-detail">{active ? <ArenaMonitor arena={active} members={members} /> : <EmptyState title="Выберите Arena" detail="Для активной Arena сообщения обновляются автоматически; завершённую можно перечитать и оставить feedback." />}</section>
    </div>
  );
}

function ArenaMonitor({ arena, members }) {
  const sessionID = arena.arena_session_id;
  const [messages, setMessages] = useState([]);
  const [session, setSession] = useState(null);
  const [feedback, setFeedback] = useState([]);
  const [error, setError] = useState("");
  const [selectedParticipant, setSelectedParticipant] = useState(arena.participants?.[0]?.user_id || "");
  const [comment, setComment] = useState("");

  useEffect(() => {
    if (!sessionID) return;
    let alive = true;
    const load = async () => {
      try {
        const [sessionData, messageData, feedbackData] = await Promise.all([api.arenaSession(sessionID), api.arenaMessages(sessionID), api.arenaFeedback(sessionID)]);
        if (!alive) return;
        setSession(sessionData); setMessages(messageData.messages || []); setFeedback(feedbackData.feedback || []); setError("");
      } catch (err) { if (alive) setError(err.message); }
    };
    load();
    const timer = setInterval(load, 2500);
    return () => { alive = false; clearInterval(timer); };
  }, [sessionID]);

  const sendFeedback = async (event) => {
    event.preventDefault();
    try {
      const result = await api.createFeedback(sessionID, { participant_id: selectedParticipant, comment, ratings: [] });
      setFeedback((items) => [...items, result]); setComment("");
    } catch (err) { setError(err.message); }
  };

  if (!sessionID) return <div className="detail-section"><p className="eyebrow">{arena.status}</p><h2>{arena.scenario_title}</h2><p>Arena Session ещё не создана. Monitor станет доступен после join участников.</p><MockBadge text="Audio/voice monitor не входит в текущий chat-only backend MVP" /></div>;

  return (
    <div className="arena-monitor">
      <header><div><p className="eyebrow">Live monitor · polling 2.5s</p><h2>{arena.scenario_title}</h2><p>{session?.status || arena.status} · v{arena.scenario_version}</p></div><span className="live-badge">● LIVE</span></header>
      <ErrorMessage>{error}</ErrorMessage>
      <div className="mvp-banner"><strong>Чат обновляется почти в real time через REST.</strong> В текущем API нет отдельного spectator WebSocket для curator; поэтому monitor использует безопасный polling истории сообщений.</div>
      <div className="chat-stream curator-stream">{messages.map((message) => <div className="chat-line" key={message.message_id}><small>{memberName(members, message.sender_id)}</small><p>{message.text}</p><time>{new Date(message.created_at).toLocaleTimeString()}</time></div>)}{!messages.length && <EmptyState title="Сообщений пока нет" detail="Как только участники начнут чат, реплики появятся здесь." />}</div>
      <div className="audio-placeholder"><div className="wave">▁▃▆▂▅▇▃▁▅▃▆▂▁</div><div><strong>Аудиозвонок / запись</strong><small>MVP mock — voice/media backend не реализован как обязательный канал</small></div><button disabled>Слушать live</button></div>
      <div className="feedback-panel">
        <h3>Экспертный feedback</h3>
        {feedback.map((item) => <article key={item.feedback_id}><strong>{memberName(members, item.participant_id)}</strong><p>{item.comment}</p><small>{new Date(item.created_at).toLocaleString()}</small></article>)}
        <form onSubmit={sendFeedback}><select required value={selectedParticipant} onChange={(e) => setSelectedParticipant(e.target.value)}>{(session?.participants || arena.participants || []).map((p) => <option key={p.user_id} value={p.user_id}>{memberName(members, p.user_id)} · {p.role}</option>)}</select><textarea required value={comment} onChange={(e) => setComment(e.target.value)} placeholder="Комментарий конкретному участнику" /><button className="primary" disabled={!comment.trim()}>Сохранить feedback</button></form>
      </div>
    </div>
  );
}

function AnalyticsView({ analytics, members }) {
  return <section className="analytics-view"><header><div><p className="eyebrow">Командная аналитика</p><h2>Срез по компетенциям</h2></div><strong className="big-metric">{members.length}<small>людей</small></strong></header>{analytics.loading && <Loading />}{analytics.error && <ErrorMessage>{analytics.error}</ErrorMessage>}<div className="analytics-grid">{analytics.data?.clusters?.map((item) => <article key={item.cluster}><strong>{item.cluster}</strong><div className="distribution"><span>0–1 <b>{item.distribution.level_0_1}</b></span><span>2–3 <b>{item.distribution.level_2_3}</b></span><span>4 <b>{item.distribution.level_4}</b></span></div></article>)}</div><div className="weak-cluster"><small>Самый частый слабый кластер</small><strong>{analytics.data?.most_common_weak_cluster || "Нет данных"}</strong></div></section>;
}

export function UserArenaHistory({ userID, history, assignments, curator = false }) {
  const [details, setDetails] = useState(null);
  const [error, setError] = useState("");
  const open = async (sessionID) => { try { setError(""); setDetails(await api.arenaHistoryDetails(userID, sessionID)); } catch (err) { setError(err.message); } };
  return <div className="detail-section"><ErrorMessage>{error}</ErrorMessage>{assignments?.data?.arenas?.length > 0 && <><h3>Запланированные</h3><div className="timeline-list">{assignments.data.arenas.map((item) => <HistoryCard key={item.arena_assignment_id} title={item.scenario_title} meta={`${new Date(item.scheduled_at).toLocaleString()} · ${item.status}`} detail="Arena assignment" />)}</div></>}<h3>История Arena</h3>{history.loading && <Loading />}{history.data?.arenas?.length ? <div className="timeline-list">{history.data.arenas.map((item) => <button className="history-card clickable" key={item.arena_session_id} onClick={() => open(item.arena_session_id)}><div><strong>{item.scenario_title}</strong><small>{item.completed_at ? new Date(item.completed_at).toLocaleString() : item.status}</small></div><span>{item.has_feedback ? "Feedback ✓" : "Без feedback"}</span></button>)}</div> : <EmptyState title="История Arena пуста" detail="Завершённые сессии появятся здесь." />}{details && <ArenaDetails details={details} curator={curator} />}</div>;
}

function ArenaDetails({ details }) {
  return <div className="arena-details"><div className="panel-title"><strong>{details.session.scenario_title} · v{details.session.scenario_version}</strong><span>{details.session.status}</span></div><div className="chat-stream compact-stream">{details.messages.map((m) => <div className="chat-line" key={m.message_id}><small>{m.sender_id}</small><p>{m.text}</p></div>)}</div><h4>Feedback</h4>{details.feedback.length ? details.feedback.map((f) => <article className="feedback-item" key={f.feedback_id}><strong>Для участника {f.participant_id}</strong><p>{f.comment}</p></article>) : <small>Feedback пока нет.</small>}{details.media?.length ? <div className="media-list">{details.media.map((media) => <a key={media.media_id} href={media.url} target="_blank" rel="noreferrer">{media.type} · {media.duration_seconds || 0} сек.</a>)}</div> : <MockBadge text="Media/audio отсутствует в этой Arena или ещё не реализовано backend" />}</div>;
}

function FeedbackOverview({ history, userID }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);
  useEffect(() => {
    const arenas = history.data?.arenas || [];
    if (!arenas.length) { setItems([]); return; }
    let alive = true; setLoading(true);
    Promise.all(arenas.filter((a) => a.has_feedback).map((a) => api.arenaHistoryDetails(userID, a.arena_session_id))).then((rows) => { if (alive) setItems(rows.flatMap((row) => row.feedback || []).filter((f) => f.participant_id === userID)); }).finally(() => alive && setLoading(false));
    return () => { alive = false; };
  }, [history.data, userID]);
  return <div className="detail-section">{loading && <Loading />}{items.length ? items.map((item) => <article className="feedback-item" key={item.feedback_id}><p>{item.comment}</p><small>{new Date(item.created_at).toLocaleString()}</small></article>) : <EmptyState title="Feedback пока нет" detail="Отзывы из завершённых Arena появятся здесь." />}</div>;
}

function HistoryCard({ title, meta, detail }) { return <article className="history-card"><div><strong>{title}</strong><small>{meta}</small></div><p>{detail}</p></article>; }
function MockBadge({ text }) { return <div className="mock-badge"><strong>MVP</strong><span>{text}</span></div>; }
function memberName(members, id) { return members.find((m) => m.user_id === id)?.name || members.find((m) => m.id === id)?.username || id; }
function dateRange(days) { const to = new Date(); const from = new Date(); from.setDate(from.getDate() - days); return { from: from.toISOString().slice(0,10), to: to.toISOString().slice(0,10) }; }
