import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getTrainingSessions } from '../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../components/UI';
import { channelLabel, runtimeLabel } from '../app/presentation';

export default function History() {
  const navigate = useNavigate(); const sessions = useQuery({ queryKey: ['training-sessions'], queryFn: ({ signal }) => getTrainingSessions(signal) });
  if (sessions.isLoading) return <LoadingState />; if (sessions.error) return <ErrorState error={sessions.error} retry={sessions.refetch} />;
  return <div className="page"><header className="page-header"><div><span className="eyebrow">Ученик</span><h1>История</h1><p>Ваши тренировки и результаты.</p></div></header>{sessions.data.length ? <div className="card-grid">{sessions.data.map((session) => <article className="card" key={session.session_id}><StatusBadge value={session.status} /><h2 style={{ marginTop: 12 }}>{session.scenario_title}</h2><p>{runtimeLabel(session.mode)} · {channelLabel(session.channel)}</p>{session.score !== null && <p><strong>{session.score} / 100</strong></p>}{session.status === 'completed' ? <button type="button" onClick={() => navigate(`/learner/training/${session.session_id}/analysis`)}>Открыть разбор</button> : <button type="button" onClick={() => navigate(`/learner/training/${session.session_id}`)}>Продолжить</button>}</article>)}</div> : <EmptyState title="История пока пуста" description="После первой тренировки здесь появится разбор." />}</div>;
}
