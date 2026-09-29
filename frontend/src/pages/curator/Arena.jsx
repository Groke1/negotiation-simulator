import { useMemo, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { createArenaAssignment, getArenaAssignments, getCuratorTeam, getScenario, getScenarios } from '../../api';
import { ClusterBars, EmptyState, ErrorState, LoadingState, StatusBadge, TagList } from '../../components/UI';

function CandidateCard({ person, selected, onSelect }) {
  return (
    <button className={selected ? 'panel candidate selected' : 'panel candidate'} type="button" onClick={() => onSelect(person)}>
      <strong>{person.name}</strong>
      <span>{person.sessions_count || 0} тренировок{person.average_score != null ? ` · score ${person.average_score}` : ''}</span>
      <ClusterBars clusters={person.clusters || []} variant="curator-clusters" showConfidence={false} />
    </button>
  );
}

export default function CuratorArena() {
  const navigate = useNavigate();
  const client = useQueryClient();
  const [query, setQuery] = useState('');
  const [participantA, setParticipantA] = useState(null);
  const [participantB, setParticipantB] = useState(null);
  const [scenarioId, setScenarioId] = useState('');
  const [schedule, setSchedule] = useState('');
  const [channel, setChannel] = useState('chat');
  const [roles, setRoles] = useState([]);
  const [roleAId, setRoleAId] = useState('');
  const [roleBId, setRoleBId] = useState('');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);
  const scheduleInputRef = useRef(null);

  const team = useQuery({ queryKey: ['curator-team'], queryFn: ({ signal }) => getCuratorTeam(signal) });
  const scenarios = useQuery({ queryKey: ['scenarios', 'arena'], queryFn: ({ signal }) => getScenarios({ status: 'published', runtime: 'arena', limit: 100 }, signal) });
  const arenas = useQuery({ queryKey: ['arena-assignments'], queryFn: ({ signal }) => getArenaAssignments({}, signal) });

  const candidates = useMemo(() => {
    const value = query.trim().toLowerCase();
    return (team.data || []).filter((person) => {
      if (person.user_id === participantA?.user_id) return false;
      return !value || person.name.toLowerCase().includes(value);
    });
  }, [team.data, participantA, query]);

  async function selectScenario(value) {
    setScenarioId(value);
    setRoles([]);
    setRoleAId('');
    setRoleBId('');
    if (!value) return;
    try {
      const detail = await getScenario(value);
      const nextRoles = detail.roles || [];
      setRoles(nextRoles);
      setRoleAId(nextRoles[0]?.id || '');
      setRoleBId(nextRoles.find((role) => role.id !== nextRoles[0]?.id)?.id || '');
    } catch (err) {
      setError(err.message);
    }
  }

  function changeRoleA(value) {
    setRoleAId(value);
    if (value === roleBId) setRoleBId(roles.find((role) => role.id !== value)?.id || '');
  }

  function changeRoleB(value) {
    setRoleBId(value);
    if (value === roleAId) setRoleAId(roles.find((role) => role.id !== value)?.id || '');
  }

  function openSchedulePicker() {
    try { scheduleInputRef.current?.showPicker?.(); } catch { /* Native picker is not available in every browser. */ }
  }

  async function submit(event) {
    event.preventDefault();
    setError('');
    if (!participantA || !participantB || !scenarioId || !schedule || !roleAId || !roleBId) {
      setError('Заполните двух участников, сценарий, разные роли и время.');
      return;
    }
    if (participantA.user_id === participantB.user_id) {
      setError('Для Arena нужны два разных участника.');
      return;
    }
    if (roleAId === roleBId) {
      setError('Участникам нельзя назначить одну и ту же переговорную роль.');
      return;
    }
    setSaving(true);
    try {
      await createArenaAssignment({
        scenario_id: scenarioId,
        participants: [
          { user_id: participantA.user_id, scenario_role_id: roleAId },
          { user_id: participantB.user_id, scenario_role_id: roleBId },
        ],
        channel,
        scheduled_at: new Date(schedule).toISOString(),
      });
      await client.invalidateQueries({ queryKey: ['arena-assignments'] });
      setParticipantA(null);
      setParticipantB(null);
      setScenarioId('');
      setSchedule('');
      setRoles([]);
      setRoleAId('');
      setRoleBId('');
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  if (team.isLoading || scenarios.isLoading || arenas.isLoading) return <LoadingState />;
  if (team.error || scenarios.error || arenas.error) return <ErrorState error={team.error || scenarios.error || arenas.error} retry={() => { team.refetch(); scenarios.refetch(); arenas.refetch(); }} />;

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <span className="eyebrow">Куратор</span>
          <h1>Арены</h1>
          <p>Подберите двух людей по профилю, выберите сценарий Arena и назначьте разные переговорные роли.</p>
        </div>
        <button className="secondary" type="button" onClick={() => navigate('/curator/team')}>Команда</button>
      </header>

      <form className="card form-grid" onSubmit={submit}>
        <h2>Новая Arena</h2>
        <label className="field">
          Поиск участников
          <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Имя сотрудника" />
        </label>
        <div className="split">
          <section className="panel">
            <h2>Участник 1</h2>
            {participantA ? (
              <>
                <strong>{participantA.name}</strong>
                <ClusterBars clusters={participantA.clusters || []} variant="curator-clusters" showConfidence={false} />
                <button className="secondary" type="button" onClick={() => setParticipantA(null)}>Изменить</button>
              </>
            ) : (
              <div className="candidate-list">
                {(team.data || []).filter((person) => !query || person.name.toLowerCase().includes(query.toLowerCase())).map((person) => <CandidateCard key={person.user_id} person={person} onSelect={setParticipantA} />)}
              </div>
            )}
          </section>
          <section className="panel">
            <h2>Участник 2</h2>
            {participantB ? (
              <>
                <strong>{participantB.name}</strong>
                <ClusterBars clusters={participantB.clusters || []} variant="curator-clusters" showConfidence={false} />
                <button className="secondary" type="button" onClick={() => setParticipantB(null)}>Изменить</button>
              </>
            ) : (
              <div className="candidate-list">
                {candidates.map((person) => <CandidateCard key={person.user_id} person={person} onSelect={setParticipantB} />)}
              </div>
            )}
          </section>
        </div>
        <label className="field">
          Сценарий
          <select value={scenarioId} onChange={(event) => selectScenario(event.target.value)}>
            <option value="">Выберите сценарий</option>
            {scenarios.data.map((scenario) => <option key={scenario.scenario_id} value={scenario.scenario_id}>{scenario.title}</option>)}
          </select>
        </label>
        {scenarioId && scenarios.data.find((scenario) => scenario.scenario_id === scenarioId) && (
          <div className="panel">
            <TagList tags={scenarios.data.find((scenario) => scenario.scenario_id === scenarioId).tags} />
            <p>{scenarios.data.find((scenario) => scenario.scenario_id === scenarioId).description}</p>
          </div>
        )}
        {roles.length === 2 && (
          <div className="split">
            <label className="field">{participantA?.name || 'Участник 1'}<select value={roleAId} onChange={(event) => changeRoleA(event.target.value)}>{roles.map((role) => <option key={role.id} value={role.id} disabled={role.id === roleBId}>{role.name}</option>)}</select></label>
            <label className="field">{participantB?.name || 'Участник 2'}<select value={roleBId} onChange={(event) => changeRoleB(event.target.value)}>{roles.map((role) => <option key={role.id} value={role.id} disabled={role.id === roleAId}>{role.name}</option>)}</select></label>
          </div>
        )}
        <label className="field arena-schedule" onClick={openSchedulePicker}>Дата и время<input ref={scheduleInputRef} type="datetime-local" required value={schedule} onClick={openSchedulePicker} onChange={(event) => setSchedule(event.target.value)} /></label>
        <label className="field">Канал<select value={channel} onChange={(event) => setChannel(event.target.value)}><option value="chat">Чат</option></select></label>
        {error && <p className="error">{error}</p>}
        <button disabled={saving}>{saving ? 'Создаём…' : 'Создать Arena'}</button>
      </form>

      <section className="card" style={{ marginTop: 16 }}>
        <h2>Все Arena</h2>
        {arenas.data.length ? (
          <div className="card-grid">
            {arenas.data.map((arena) => (
              <article className="panel scenario-link-card" key={arena.arena_assignment_id} role={arena.arena_session_id ? 'link' : undefined} tabIndex={arena.arena_session_id ? 0 : undefined} onClick={() => arena.arena_session_id && navigate(`/curator/arenas/${arena.arena_session_id}`)} onKeyDown={(event) => { if (arena.arena_session_id && (event.key === 'Enter' || event.key === ' ')) navigate(`/curator/arenas/${arena.arena_session_id}`); }}>
                <StatusBadge value={arena.status} />
                <h2 style={{ marginTop: 12 }}>{arena.scenario_title}</h2>
                <p>{new Date(arena.scheduled_at).toLocaleString('ru-RU')}</p>
                {arena.arena_session_id && <button type="button" className="secondary" onClick={(event) => { event.stopPropagation(); navigate(`/curator/arenas/${arena.arena_session_id}`); }}>{arena.status === 'completed' ? 'Открыть feedback' : 'Открыть детали'}</button>}
              </article>
            ))}
          </div>
        ) : <EmptyState title="Arena пока не назначены" />}
      </section>
    </div>
  );
}
