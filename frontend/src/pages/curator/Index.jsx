import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getCuratorAnalytics, getCuratorTeam } from '../../api';
import { ClusterBars, EmptyState, ErrorState, LoadingState } from '../../components/UI';

export default function Curator() {
  const navigate = useNavigate();
  const team = useQuery({ queryKey: ['curator-team'], queryFn: ({ signal }) => getCuratorTeam(signal) });
  const analytics = useQuery({ queryKey: ['curator-analytics'], queryFn: ({ signal }) => getCuratorAnalytics(signal) });

  if (team.isLoading) return <LoadingState />;
  if (team.error) return <ErrorState error={team.error} retry={team.refetch} />;

  const summary = analytics.data && [
    ['Ученики', team.data.length],
    ['Тренировки', analytics.data.training_sessions_count],
    ['Arena', analytics.data.arena_sessions_count],
    ['Ожидают', analytics.data.pending_assignments_count ?? '—'],
  ];

  return <div className="page">
    <header className="page-header"><div><span className="eyebrow">Куратор</span><h1>Команда</h1><p>Назначайте практики, наблюдайте прогресс и создавайте живые Arena.</p></div><button onClick={() => navigate('/curator/arenas')}>Создать Arena</button></header>
    {summary && <section className="card" style={{ display: 'grid', gridTemplateColumns: 'repeat(4, minmax(0, 1fr))', gap: 24, marginBottom: 16 }} aria-label="Статистика команды">{summary.map(([label, value]) => <div key={label}><span className="eyebrow">{label}</span><h2 style={{ margin: '6px 0 0', fontSize: 22 }}>{value}</h2></div>)}</section>}
    {team.data.length ? <div className="card-grid">{team.data.map((member) => <article className="card scenario-link-card" key={member.user_id} role="link" tabIndex="0" onClick={() => navigate(`/curator/team/${member.user_id}`)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') navigate(`/curator/team/${member.user_id}`); }}><h2>{member.name}</h2><p>{member.sessions_count} тренировок{member.last_session_at && ` · последняя ${new Date(member.last_session_at).toLocaleDateString('ru-RU')}`}</p>{member.average_score !== null && <p><strong>{member.average_score}</strong> средний балл</p>}<ClusterBars clusters={member.clusters || []} variant="curator-clusters" showConfidence={false} /></article>)}</div> : <EmptyState title="В команде пока нет учеников" description="Когда участники присоединятся к организации, они появятся здесь." />}
  </div>;
}
