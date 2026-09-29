import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getMyArenaAssignments, getMyArenaHistory } from '../../api';
import { useAuth } from '../../app/AuthContext';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../../components/UI';

const formatDateTime = (value) => value ? new Date(value).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' }) : 'время уточняется';

function navigateWithKeyboard(event, onOpen) {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault();
    onOpen();
  }
}

export default function ArenaIndex() {
  const { me } = useAuth();
  const navigate = useNavigate();
  const upcoming = useQuery({ queryKey: ['my-arena-assignments', me?.id], queryFn: ({ signal }) => getMyArenaAssignments(me.id, signal), enabled: Boolean(me?.id) });
  const history = useQuery({ queryKey: ['my-arena-history', me?.id], queryFn: ({ signal }) => getMyArenaHistory(me.id, signal), enabled: Boolean(me?.id) });
  const upcomingArenas = (upcoming.data || []).filter((arena) => ['scheduled', 'in_progress'].includes(arena.status));
  const completedArenas = (history.data || []).filter((arena) => arena.status === 'completed');

  if (upcoming.isLoading || history.isLoading) return <LoadingState />;
  if (upcoming.error || history.error) return <ErrorState error={upcoming.error || history.error} retry={upcoming.error ? upcoming.refetch : history.refetch} />;

  return <div className="page">
    <header className="page-header"><div><span className="eyebrow">Ученик</span><h1>Арены</h1><p>Живые переговоры с другим участником — без оценок и подсказок в чате.</p></div></header>
    <section className="card" style={{ marginBottom: 16 }}>
      <h2>Предстоящие</h2>
      {upcomingArenas.length ? <div className="card-grid">{upcomingArenas.map((arena) => {
        const open = () => navigate(`/learner/arenas/join/${arena.arena_assignment_id}`);
        return <article className="panel scenario-link-card arena-upcoming-card" key={arena.arena_assignment_id} role="link" tabIndex="0" onClick={open} onKeyDown={(event) => navigateWithKeyboard(event, open)}>
          <StatusBadge value={arena.status} />
          <h2>{arena.scenario_title}</h2>
          <p>{formatDateTime(arena.scheduled_at)}</p>
          <p className="card-link-hint">{arena.status === 'in_progress' ? 'Продолжить переговоры →' : 'Открыть приглашение →'}</p>
        </article>;
      })}</div> : <EmptyState title="Ближайших Arena нет" description="Куратор создаст приглашение, когда подберёт пару и сценарий." />}
    </section>
    <section className="card">
      <h2>История</h2>
      {completedArenas.length ? <div className="card-grid">{completedArenas.map((arena) => {
        const open = () => navigate(`/learner/arenas/${arena.arena_session_id}`);
        const finishedAt = arena.completed_at || arena.started_at || arena.scheduled_at;
        return <article className="panel scenario-link-card arena-history-card" key={arena.arena_session_id} role="link" tabIndex="0" onClick={open} onKeyDown={(event) => navigateWithKeyboard(event, open)}>
          <header><StatusBadge value={arena.status} />{arena.has_feedback && <span className="arena-feedback-badge">Feedback готов</span>}</header>
          <span className="eyebrow">Живые переговоры</span>
          <h2>{arena.scenario_title}</h2>
          <dl>
            <div><dt>Завершена</dt><dd>{formatDateTime(finishedAt)}</dd></div>
            <div><dt>Итог</dt><dd>{arena.has_feedback ? 'Есть feedback куратора' : 'Диалог доступен для просмотра'}</dd></div>
          </dl>
          <p className="card-link-hint">Открыть диалог и результат →</p>
        </article>;
      })}</div> : <EmptyState title="Завершённых Arena пока нет" description="После первой завершённой Arena здесь появятся диалог и результат." />}
    </section>
  </div>;
}
