import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getAssignments, getCuratorTeamSessions } from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../../components/UI';

export default function CuratorHistory({ assignments = false }) {
  const navigate = useNavigate();
  const [type, setType] = useState('');
  const [status, setStatus] = useState('');
  const filters = assignments ? { status: status || undefined, limit: 100 } : { type: type || undefined, limit: 100 };
  const source = useQuery({
    queryKey: [assignments ? 'assignments' : 'curator-history', filters],
    queryFn: ({ signal }) => assignments ? getAssignments(filters, signal) : getCuratorTeamSessions(filters, signal),
  });

  if (source.isLoading) return <LoadingState />;
  if (source.error) return <ErrorState error={source.error} retry={source.refetch} />;

  const items = source.data || [];
  const openItem = (item) => {
    if (assignments) return;
    if (item.type === 'arena') navigate(`/curator/arenas/${item.session_id}`);
    else if (item.user_ids?.[0]) navigate(`/curator/team/${item.user_ids[0]}`);
  };
  return (
    <div className="page">
      <header className="page-header">
        <div>
          <span className="eyebrow">Куратор</span>
          <h1>{assignments ? 'Назначения' : 'История команды'}</h1>
          <p>{assignments ? 'Активные и завершённые назначения тренировок.' : 'Тренировочные и Arena-сессии команды.'}</p>
        </div>
      </header>
      {assignments ? (
        <div className="tabs">
          {[
            ['', 'Все'],
            ['pending', 'Ожидают'],
            ['completed', 'Завершены'],
          ].map(([value, label]) => <button key={label} className={status === value ? 'active' : ''} type="button" onClick={() => setStatus(value)}>{label}</button>)}
        </div>
      ) : (
        <div className="tabs">
          {[
            ['', 'Все'],
            ['training', 'Тренировки'],
            ['arena', 'Arena'],
          ].map(([value, label]) => <button key={label} className={type === value ? 'active' : ''} type="button" onClick={() => setType(value)}>{label}</button>)}
        </div>
      )}
      {items.length ? (
        <div className="card-grid">
          {items.map((item) => (
            <article className={`card ${!assignments ? 'scenario-link-card' : ''}`} key={item.assignment_id || item.session_id} role={!assignments ? 'link' : undefined} tabIndex={!assignments ? 0 : undefined} onClick={() => openItem(item)} onKeyDown={(event) => { if (!assignments && (event.key === 'Enter' || event.key === ' ')) openItem(item); }}>
              <StatusBadge value={item.status} />
              <h2 style={{ marginTop: 12 }}>{item.scenario_title}</h2>
              {assignments ? (
                <p>{item.status === 'pending' ? 'Ожидает прохождения' : 'Завершено'} · {item.created_at ? new Date(item.created_at).toLocaleDateString('ru-RU') : '—'}</p>
              ) : (
                <><p>{item.type === 'arena' ? 'Arena' : 'Тренировка'} · {Number(item.score) >= 0 ? item.score : 'без оценки'}</p><p className="card-link-hint">{item.type === 'arena' ? 'Открыть диалог и feedback' : 'Открыть профиль ученика'}</p></>
              )}
            </article>
          ))}
        </div>
      ) : <EmptyState title={assignments ? 'Назначений нет' : 'История пока пуста'} />}
    </div>
  );
}
