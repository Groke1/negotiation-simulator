import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { createOrganization, getCompetencyClusters, getMyProfile, getOrganizations, getProfileHistory, setOrganization } from '../api';
import { EmptyState, ErrorState, LoadingState } from '../components/UI';
import ProfileMetricsCard from '../components/ProfileMetricsCard';
import { useAuth } from '../app/AuthContext';
import { formatDelta, formatNumber } from '../app/format';

const sourceLabel = (source) => source === 'arena_feedback' ? 'Arena · feedback куратора' : 'Тренировка';

export default function Profile() {
  const { me, signInComplete } = useAuth();
  const navigate = useNavigate();
  const profile = useQuery({ queryKey: ['my-profile'], queryFn: ({ signal }) => getMyProfile(signal) });
  const clusters = useQuery({ queryKey: ['competency-clusters'], queryFn: ({ signal }) => getCompetencyClusters(signal), staleTime: 60 * 60 * 1000 });
  const history = useQuery({ queryKey: ['my-profile-history', me?.id], queryFn: ({ signal }) => getProfileHistory(me.id, signal), enabled: Boolean(me?.id) });
  const organizations = useQuery({ queryKey: ['organizations'], queryFn: ({ signal }) => getOrganizations(signal), enabled: !me?.organization });
  const [organizationId, setOrganizationId] = useState('');
  const [name, setName] = useState('');
  const [message, setMessage] = useState('');
  const [saving, setSaving] = useState(false);

  async function choose(event) {
    event.preventDefault(); setSaving(true); setMessage('');
    try { await setOrganization(organizationId); await signInComplete(); setMessage('Организация выбрана.'); } catch (requestError) { setMessage(requestError.message); } finally { setSaving(false); }
  }
  async function create(event) {
    event.preventDefault(); setSaving(true); setMessage('');
    try { const organization = await createOrganization(name.trim()); await setOrganization(organization.id); await signInComplete(); setMessage('Организация создана и выбрана.'); } catch (requestError) { setMessage(requestError.message); } finally { setSaving(false); }
  }
  function openHistoryItem(event) {
    navigate(event.source === 'arena_feedback' ? `/learner/arenas/${event.source_id}` : `/learner/training/${event.source_id}/analysis`);
  }

  if (profile.isLoading || clusters.isLoading || history.isLoading) return <LoadingState />;
  if (profile.error || clusters.error || history.error) return <ErrorState error={profile.error || clusters.error || history.error} retry={() => { profile.refetch(); clusters.refetch(); history.refetch(); }} />;
  return <div className="page">
    <header className="page-header"><div><span className="eyebrow">Ученик</span><h1>Профиль</h1><p>{me?.email}</p></div></header>
    {!me?.organization && <section className="card" style={{ marginBottom: 16 }}><h2>Завершите настройку</h2><p>Выберите организацию, чтобы открыть рабочее пространство.</p>{organizations.data?.length ? <form className="form-grid" onSubmit={choose}><label className="field">Организация<select required value={organizationId} onChange={(event) => setOrganizationId(event.target.value)}><option value="">Выберите организацию</option>{organizations.data.map((organization) => <option key={organization.id} value={organization.id}>{organization.name}</option>)}</select></label><button disabled={saving || !organizationId}>Продолжить</button></form> : null}<form className="form-grid" onSubmit={create} style={{ marginTop: 16 }}><label className="field">Название новой организации<input required value={name} onChange={(event) => setName(event.target.value)} /></label><button className="secondary" disabled={saving || !name.trim()}>Создать организацию</button></form>{message && <p className={message.includes('Ошибка') ? 'error' : 'success'}>{message}</p>}</section>}
    <ProfileMetricsCard clusters={profile.data?.clusters || []} dictionary={clusters.data || []} className="profile-metrics-only" />
    <section className="card profile-history" style={{ marginTop: 16 }}><h2>Изменения профиля</h2>{history.data?.events?.length ? <div className="form-grid">{history.data.events.map((event) => <button type="button" className="panel profile-history-card" key={`${event.source}-${event.source_id}`} onClick={() => openHistoryItem(event)}><span className="eyebrow">{sourceLabel(event.source)}</span><time>{event.created_at ? new Date(event.created_at).toLocaleString('ru-RU') : ''}</time>{event.changes.map((change) => <span key={change.cluster_id}><strong>{clusters.data.find((cluster) => cluster.id === change.cluster_id)?.name || 'Компетенция'}:</strong> {formatNumber(change.before)} → {formatNumber(change.after)} ({formatDelta(change.delta)})</span>)}<small>Открыть подробный отчёт →</small></button>)}</div> : <EmptyState title="Изменений профиля пока нет" description="После завершённой тренировки или feedback по Arena здесь появится динамика." />}</section>
  </div>;
}
