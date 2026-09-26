import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { createAssignment, getCompetencyClusters, getProfileHistory, getScenarios, getUserProfile } from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge, TagList } from '../../components/UI';
import ProfileMetricsCard from '../../components/ProfileMetricsCard';
import { formatDelta, formatNumber } from '../../app/format';

export default function Learner() {
  const { userId } = useParams();
  const navigate = useNavigate();
  const client = useQueryClient();
  const [assign, setAssign] = useState(false);
  const [selected, setSelected] = useState('');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const profile = useQuery({ queryKey: ['profile', userId], queryFn: ({ signal }) => getUserProfile(userId, signal) });
  const profileHistory = useQuery({ queryKey: ['profile-history', userId], queryFn: ({ signal }) => getProfileHistory(userId, signal) });
  const scenarios = useQuery({ queryKey: ['scenarios', 'assignment'], queryFn: ({ signal }) => getScenarios({ status: 'published', limit: 100 }, signal) });
  const dictionary = useQuery({ queryKey: ['competency-clusters'], queryFn: ({ signal }) => getCompetencyClusters(signal), staleTime: 60 * 60 * 1000 });
  const clusterNames = new Map((dictionary.data || []).map((cluster) => [cluster.id, cluster.name]));

  async function submit() {
    if (!selected) return;
    setSaving(true);
    setError('');
    try {
      await createAssignment(userId, selected);
      await client.invalidateQueries({ queryKey: ['assignments'] });
      setAssign(false);
      setSelected('');
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSaving(false);
    }
  }

  function openHistoryItem(event) {
    navigate(event.source === 'arena_feedback' ? `/curator/arenas/${event.source_id}` : `/curator/training/${event.source_id}/analysis`);
  }

  if (profile.isLoading || dictionary.isLoading || profileHistory.isLoading) return <LoadingState />;
  if (profile.error || dictionary.error || profileHistory.error) return <ErrorState error={profile.error || dictionary.error || profileHistory.error} retry={() => { profile.refetch(); dictionary.refetch(); profileHistory.refetch(); }} />;

  return <div className="page">
    <header className="page-header"><div><span className="eyebrow">Куратор · ученик</span><h1>{profile.data.name || 'Ученик'}</h1><p>Профиль развивается только по результатам сессий, без клиентских выводов.</p></div><div><button className="secondary" type="button" onClick={() => navigate('/curator/team')}>Назад</button><button type="button" onClick={() => setAssign(true)}>Назначить тренировку</button></div></header>
    <ProfileMetricsCard clusters={profile.data.clusters || []} dictionary={dictionary.data} />
    <section className="card" style={{ marginTop: 16 }}><h2>Изменения профиля</h2>{profileHistory.data?.events?.length ? <div className="form-grid">{profileHistory.data.events.map((event) => <button className="panel profile-history-card" type="button" key={`${event.source}-${event.source_id}`} onClick={() => openHistoryItem(event)}><span className="eyebrow">{event.source === 'arena_feedback' ? 'Arena · feedback куратора' : 'Тренировка'}</span><time>{event.created_at ? new Date(event.created_at).toLocaleString('ru-RU') : ''}</time>{event.changes.map((change) => <span key={change.cluster_id}><strong>{clusterNames.get(change.cluster_id) || 'Компетенция'}:</strong> {formatNumber(change.before)} → {formatNumber(change.after)} ({formatDelta(change.delta)})</span>)}<small>Открыть подробный отчёт →</small></button>)}</div> : <EmptyState title="Изменений профиля пока нет" description="После завершённой тренировки или feedback по Arena здесь появится динамика." />}</section>
    {assign && <div className="modal-overlay"><div className="modal"><h2>Назначить тренировку</h2><p>Выберите опубликованный сценарий для ученика.</p>{scenarios.isLoading ? <LoadingState /> : scenarios.data?.length ? <div className="form-grid">{scenarios.data.map((scenario) => <label className="panel" key={scenario.scenario_id}><input type="radio" name="scenario" value={scenario.scenario_id} checked={selected === scenario.scenario_id} onChange={() => setSelected(scenario.scenario_id)} /> <strong>{scenario.title}</strong><p>{scenario.description}</p><TagList tags={scenario.tags} /><StatusBadge value={scenario.difficulty} /></label>)}</div> : <EmptyState title="Нет опубликованных сценариев" />}{error && <p className="error">{error}</p>}<button type="button" onClick={submit} disabled={!selected || saving}>{saving ? 'Назначаем…' : 'Назначить'}</button><button className="secondary" type="button" onClick={() => setAssign(false)}>Отмена</button></div></div>}
  </div>;
}
