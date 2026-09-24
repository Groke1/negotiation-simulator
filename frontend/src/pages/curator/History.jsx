import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getAssignments, getCuratorTeamSessions } from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../../components/UI';

export default function CuratorHistory({ assignments = false }) {
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
            <article className="card" key={item.assignment_id || item.session_id}>
              <StatusBadge value={item.status} />
              <h2 style={{ marginTop: 12 }}>{item.scenario_title}</h2>
              {assignments ? (
                <p>{item.status === 'pending' ? 'Ожидает прохождения' : 'Завершено'} · {item.created_at ? new Date(item.created_at).toLocaleDateString('ru-RU') : '—'}</p>
              ) : (
                <p>{item.type === 'arena' ? 'Arena' : 'Тренировка'} · {item.score ?? 'без score'}</p>
              )}
            </article>
          ))}
        </div>
      ) : <EmptyState title={assignments ? 'Назначений нет' : 'История пока пуста'} />}
    </div>
  );
}
