import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getArenaAssignments, getAssignments, getCuratorTeamSessions, getOrganizationUsers } from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../../components/UI';

const sessionDate = (item) => item.completed_at || item.started_at || item.scheduled_at || item.created_at || '';
const byNewest = (left, right) => Date.parse(sessionDate(right)) - Date.parse(sessionDate(left));

function arenaAsHistoryItem(arena) {
  return {
    type: 'arena',
    session_id: arena.arena_session_id,
    assignment_id: arena.arena_assignment_id,
    user_ids: (arena.participants || []).map((participant) => participant.user_id),
    scenario_title: arena.scenario_title,
    status: arena.status,
    scheduled_at: arena.scheduled_at,
    created_at: arena.created_at,
  };
}

function UserLink({ userId, name, onOpen }) {
  if (!userId) return <span>{name}</span>;
  return <button className="text-button card-user-link" type="button" onClick={(event) => { event.stopPropagation(); onOpen(userId); }}>{name}</button>;
}

export default function CuratorHistory({ assignments = false }) {
  const navigate = useNavigate();
  const [type, setType] = useState('');
  const [status, setStatus] = useState('');
  const [notice, setNotice] = useState('');
  const assignmentsQuery = useQuery({
    queryKey: ['assignments', status],
    queryFn: async ({ signal }) => {
      const statuses = status ? [status] : ['pending', 'completed'];
      const responses = await Promise.all(statuses.map((value) => getAssignments({ status: value, limit: 100 }, signal)));
      return Array.from(new Map(responses.flat().map((item) => [item.assignment_id, item])).values()).sort(byNewest);
    },
    enabled: assignments,
  });
  const training = useQuery({ queryKey: ['curator-training-history'], queryFn: ({ signal }) => getCuratorTeamSessions({ type: 'training', limit: 100 }, signal), enabled: !assignments && type !== 'arena' });
  const arenas = useQuery({ queryKey: ['curator-arena-history'], queryFn: ({ signal }) => getArenaAssignments({ limit: 100 }, signal), enabled: !assignments && type !== 'training' });
  const users = useQuery({ queryKey: ['organization-users', 'curator-history'], queryFn: ({ signal }) => getOrganizationUsers({}, signal) });
  const userNames = useMemo(() => new Map((users.data || []).map((user) => [user.id, user.username || user.name || user.email])), [users.data]);
  const items = useMemo(() => {
    if (assignments) return assignmentsQuery.data || [];
    const trainingItems = (training.data || []).filter((item) => item.type !== 'arena');
    const arenaItems = (arenas.data || []).map(arenaAsHistoryItem);
    if (type === 'training') return trainingItems.sort(byNewest);
    if (type === 'arena') return arenaItems.sort(byNewest);
    return [...trainingItems, ...arenaItems].sort(byNewest);
  }, [arenas.data, assignments, assignmentsQuery.data, training.data, type]);
  const loading = assignments ? assignmentsQuery.isLoading || users.isLoading : training.isLoading || arenas.isLoading || users.isLoading;
  const error = assignments ? assignmentsQuery.error || users.error : training.error || arenas.error || users.error;
  const retry = () => { users.refetch(); if (assignments) assignmentsQuery.refetch(); else { training.refetch(); arenas.refetch(); } };
  if (loading) return <LoadingState />;
  if (error) return <ErrorState error={error} retry={retry} />;

  const openUser = (userId) => navigate(`/curator/team/${userId}`);
  const openItem = (item) => {
    if (assignments) return openUser(item.learner_id);
    if (item.type === 'arena') return navigate(item.session_id ? `/curator/arenas/${item.session_id}` : '/curator/arenas');
    if (item.status !== 'completed') {
      setNotice('Тренировка ещё не завершена. Разбор появится после её автоматического завершения.');
      return;
    }
    return navigate(`/curator/training/${item.session_id}/analysis`);
  };
  const emptyTitle = assignments ? 'Назначений нет' : type === 'arena' ? 'Arena-сессий пока нет' : type === 'training' ? 'Тренировок пока нет' : 'История пока пуста';

  return <div className="page">
    <header className="page-header"><div><span className="eyebrow">Куратор</span><h1>{assignments ? 'Назначения' : 'История команды'}</h1><p>{assignments ? 'Активные и завершённые назначения тренировок.' : 'Тренировочные и Arena-сессии команды.'}</p></div></header>
    {assignments ? <div className="tabs">{[['', 'Все'], ['pending', 'Ожидают'], ['completed', 'Завершены']].map(([value, label]) => <button key={label} className={status === value ? 'active' : ''} type="button" onClick={() => setStatus(value)}>{label}</button>)}</div> : <div className="tabs">{[['', 'Все'], ['training', 'Тренировки'], ['arena', 'Arena']].map(([value, label]) => <button key={label} className={type === value ? 'active' : ''} type="button" onClick={() => setType(value)}>{label}</button>)}</div>}
    {notice && <p className="feedback" role="status">{notice}</p>}
    {items.length ? <div className="card-grid">{items.map((item) => {
      const participantNames = (item.user_ids || []).map((id) => userNames.get(id) || 'Ученик');
      const learnerName = userNames.get(item.learner_id) || 'Ученик';
      return <article className="card scenario-link-card" key={item.assignment_id || `${item.type}-${item.session_id}`} role="link" tabIndex="0" onClick={() => openItem(item)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') openItem(item); }}>
        <StatusBadge value={item.status} />
        <h2 style={{ marginTop: 12 }}>{item.scenario_title}</h2>
        {assignments ? <><p>Назначено: <UserLink userId={item.learner_id} name={learnerName} onOpen={openUser} /></p><p>{item.status === 'pending' ? 'Ожидает прохождения' : 'Завершено'} · {item.created_at ? new Date(item.created_at).toLocaleDateString('ru-RU') : '—'}</p><p className="card-link-hint">Открыть профиль ученика</p></> : item.type === 'arena' ? <><p>Arena · {item.scheduled_at ? new Date(item.scheduled_at).toLocaleDateString('ru-RU') : 'дата уточняется'}</p><p>Участники: {participantNames.join(' · ') || 'не указаны'}</p><p className="card-link-hint">{item.session_id ? 'Открыть диалог и feedback' : 'Открыть Arena'}</p></> : <><p>Тренировка · {Number(item.score) >= 0 ? item.score : 'без оценки'}</p><p>Ученик: <UserLink userId={item.user_ids?.[0]} name={participantNames[0] || 'Ученик'} onOpen={openUser} /></p><p className="card-link-hint">{item.status === 'completed' ? 'Открыть разбор тренировки' : 'Тренировка ещё не завершена'}</p></>}
      </article>;
    })}</div> : <EmptyState title={emptyTitle} />}
  </div>;
}
