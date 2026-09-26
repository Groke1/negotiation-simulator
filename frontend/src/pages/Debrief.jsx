import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getCompetencyClusters, getTrainingAnalysis } from '../api';
import { ClusterBars, EmptyState, ErrorState, LoadingState } from '../components/UI';

export default function Debrief() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const analysis = useQuery({ queryKey: ['training-analysis', sessionId], queryFn: ({ signal }) => getTrainingAnalysis(sessionId, signal) });
  const dictionary = useQuery({ queryKey: ['competency-clusters'], queryFn: ({ signal }) => getCompetencyClusters(signal), staleTime: 60 * 60 * 1000 });
  if (analysis.isLoading || dictionary.isLoading) return <LoadingState />;
  if (analysis.error || dictionary.error) return <ErrorState error={analysis.error || dictionary.error} retry={() => { analysis.refetch(); dictionary.refetch(); }} />;
  const data = analysis.data;
  const result = data.result || {};
  const clusterNames = new Map((dictionary.data || []).map((cluster) => [cluster.id, cluster.name]));
  return <div className="page">
    <header className="page-header"><div><span className="eyebrow">Разбор тренировки</span><h1>{Math.floor(Number(data.score || 0))} / 100</h1><p>{result.summary || 'Итог тренировки сформирован.'}</p></div><div><button type="button" onClick={() => navigate('/learner/training')}>К тренировкам</button></div></header>
    {(result.goals_met?.length || result.goals_missed?.length) ? <section className="card" style={{ marginBottom: 16 }}><h2>Результат переговоров</h2><div className="split">{result.goals_met?.length ? <div><span className="eyebrow">Достигнуто</span><ul>{result.goals_met.map((goal) => <li key={goal}>{goal}</li>)}</ul></div> : <div><span className="eyebrow">Достигнуто</span><p>Достигнутые цели не отмечены.</p></div>}{result.goals_missed?.length ? <div><span className="eyebrow">Не достигнуто</span><ul>{result.goals_missed.map((goal) => <li key={goal}>{goal}</li>)}</ul></div> : <div><span className="eyebrow">Не достигнуто</span><p>Недостигнутые цели не отмечены.</p></div>}</div></section> : null}
    <div className="split"><section className="card"><h2>Что получилось</h2>{data.strengths?.length ? data.strengths.map((item) => <p className="feedback good" key={item.title}>✓ <strong>{item.title}</strong><br />{item.description}</p>) : <p>Сильные стороны пока не сформированы.</p>}</section><section className="card"><h2>Что улучшить</h2>{data.improvements?.length ? data.improvements.map((item) => <p className="feedback bad" key={item.title}>! <strong>{item.title}</strong><br />{item.description}</p>) : <p>Рекомендаций пока нет.</p>}</section></div>
    <section className="card" style={{ marginTop: 16 }}><h2>Компетенции в попытке</h2><ClusterBars clusters={data.competencies} dictionary={dictionary.data} />{data.competencies?.some((item) => item.comment) && <div className="form-grid" style={{ marginTop: 16 }}>{data.competencies.filter((item) => item.comment).map((item) => <p className="panel" key={item.cluster_id}><strong>{clusterNames.get(item.cluster_id) || 'Компетенция'}</strong><br />{item.comment}</p>)}</div>}</section>
    <section className="card" style={{ marginTop: 16 }}><h2>Как развивался диалог</h2>{data.process?.length ? <ol className="form-grid">{data.process.map((item) => <li className="panel" key={item.order}><strong>Шаг {item.order}</strong><p>{item.summary}</p></li>)}</ol> : <EmptyState title="Ход диалога пока не сформирован" />}</section>
    <section className="card" style={{ marginTop: 16 }}><h2>Изменение профиля</h2><ClusterBars clusters={data.profile_change?.changes?.map((item) => ({ cluster_id: item.cluster_id, score: item.after })) || []} dictionary={dictionary.data} /></section>
  </div>;
}
