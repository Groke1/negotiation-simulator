import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getArenaAssignments, getAssignments, getCuratorTeamSessions } from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../../components/UI';

const sessionDate = (item) => item.completed_at || item.started_at || item.scheduled_at || item.created_at || '';
const byNewest = (left, right) => Date.parse(sessionDate(right)) - Date.parse(sessionDate(left));

function arenaAsHistoryItem(arena) {
  return {
    type: 'arena', session_id: arena.arena_session_id, assignment_id: arena.arena_assignment_id,
    user_ids: (arena.participants || []).map((participant) => participant.user_id), scenario_title: arena.scenario_title,
    status: arena.status, scheduled_at: arena.scheduled_at, created_at: arena.created_at,
  };
}

export default function CuratorHistory({ assignments = false }) {
  const navigate = useNavigate();
  const [type, setType] = useState('');
  const [status, setStatus] = useState('');
  const assignmentsQuery = useQuery({ queryKey: ['assignments', status], queryFn: ({ signal }) => getAssignments({ status: status || undefined, limit: 100 }, signal), enabled: assignments });
  const training = useQuery({ queryKey: ['curator-training-history'], queryFn: ({ signal }) => getCuratorTeamSessions({ type: 'training', limit: 100 }, signal), enabled: !assignments && type !== 'arena' });
  const arenas = useQuery({ queryKey: ['curator-arena-history'], queryFn: ({ signal }) => getArenaAssignments({ limit: 100 }, signal), enabled: !assignments && type !== 'training' });
  const items = useMemo(() => {
    if (assignments) return assignmentsQuery.data || [];
    const trainingItems = (training.data || []).filter((item) => item.type !== 'arena');
    const arenaItems = (arenas.data || []).map(arenaAsHistoryItem);
    if (type === 'training') return trainingItems.sort(byNewest);
    if (type === 'arena') return arenaItems.sort(byNewest);
    return [...trainingItems, ...arenaItems].sort(byNewest);
  }, [arenas.data, assignments, assignmentsQuery.data, training.data, type]);
  const loading = assignments ? assignmentsQuery.isLoading : training.isLoading || arenas.isLoading;
  const error = assignments ? assignmentsQuery.error : training.error || arenas.error;
  const retry = () => { if (assignments) assignmentsQuery.refetch(); else { training.refetch(); arenas.refetch(); } };
  if (loading) return <LoadingState />;
  if (error) return <ErrorState error={error} retry={retry} />;

  const openItem = (item) => {
    if (assignments) { if (item.learner_id) navigate(`/curator/team/${item.learner_id}`); return; }
    if (item.type === 'arena') { navigate(item.session_id ? `/curator/arenas/${item.session_id}` : '/curator/arenas'); return; }
    if (item.user_ids?.[0]) navigate(`/curator/team/${item.user_ids[0]}`);
  };
  const emptyTitle = assignments ? 'Назначений нет' : type === 'arena' ? 'Arena-сессий пока нет' : type === 'training' ? 'Тренировок пока нет' : 'История пока пуста';
  return <div className="page"><header className="page-header"><div><span className="eyebrow">Куратор</span><h1>{assignments ? 'Назначения' : 'История команды'}</h1><p>{assignments ? 'Активные и завершённые назначения тренировок.' : 'Тренировочные и Arena-сессии команды.'}</p></div></header>{assignments ? <div className="tabs">{[['', 'Все'], ['pending', 'Ожидают'], ['completed', 'Завершены']].map(([value, label]) => <button key={label} className={status === value ? 'active' : ''} type="button" onClick={() => setStatus(value)}>{label}</button>)}</div> : <div className="tabs">{[['', 'Все'], ['training', 'Тренировки'], ['arena', 'Arena']].map(([value, label]) => <button key={label} className={type === value ? 'active' : ''} type="button" onClick={() => setType(value)}>{label}</button>)}</div>}{items.length ? <div className="card-grid">{items.map((item) => <article className="card scenario-link-card" key={item.assignment_id || `${item.type}-${item.session_id}`} role="link" tabIndex="0" onClick={() => openItem(item)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') openItem(item); }}><StatusBadge value={item.status} /><h2 style={{ marginTop: 12 }}>{item.scenario_title}</h2>{assignments ? <p>{item.status === 'pending' ? 'Ожидает прохождения' : 'Завершено'} · {item.created_at ? new Date(item.created_at).toLocaleDateString('ru-RU') : '—'}</p> : <><p>{item.type === 'arena' ? 'Arena' : 'Тренировка'} · {item.type === 'arena' ? (item.scheduled_at ? new Date(item.scheduled_at).toLocaleDateString('ru-RU') : 'дата уточняется') : (Number(item.score) >= 0 ? item.score : 'без оценки')}</p><p className="card-link-hint">{item.type === 'arena' ? (item.session_id ? 'Открыть диалог и feedback' : 'Открыть Arena') : 'Открыть профиль ученика'}</p></>}</article>)}</div> : <EmptyState title={emptyTitle} />}</div>;
}
