import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { createOrganization, getOrganizations, grantRole, setOrganization } from '../api';
import { useAuth } from '../app/AuthContext';
import { preferredRole, roleHome, roleName } from '../app/roles';
import { ErrorState, LoadingState } from '../components/UI';

const roles = ['learner', 'methodist', 'curator'];
const descriptions = { learner: 'Проходит тренировки и участвует в Arena.', methodist: 'Создаёт сценарии переговоров.', curator: 'Назначает практики и анализирует результаты.' };

export default function Onboarding() {
  const { me, signInComplete, setRole } = useAuth();
  const organizations = useQuery({ queryKey: ['organizations'], queryFn: ({ signal }) => getOrganizations(signal), enabled: !me?.organization });
  const [organizationId, setOrganizationId] = useState(''); const [organizationName, setOrganizationName] = useState(''); const [selectedRoles, setSelectedRoles] = useState(me?.roles || []); const [error, setError] = useState(''); const [saving, setSaving] = useState(false);
  if (!me) return <LoadingState />;
  if (me.organization && me.roles?.length) return <Navigate to={roleHome(me.active_role)} replace />;
  const toggleRole = (role) => setSelectedRoles((current) => current.includes(role) ? current.filter((item) => item !== role) : [...current, role]);
  async function complete() {
    setSaving(true); setError('');
    try {
      let id = organizationId;
      if (!me.organization && !id) { const organization = await createOrganization(organizationName.trim()); id = organization.id; }
      if (!me.organization) await setOrganization(id);
      for (const role of selectedRoles.filter((value) => !me.roles.includes(value))) await grantRole(role);
      // POST /roles only persists the role. Switching the active role is what issues
      // a fresh access token with the newly granted permissions.
      await setRole(preferredRole(selectedRoles));
      await signInComplete();
    } catch (reason) { setError(reason.message); } finally { setSaving(false); }
  }
  if (organizations.error) return <ErrorState error={organizations.error} retry={organizations.refetch} />;
  const cannotContinue = saving || selectedRoles.length === 0 || (!me.organization && !organizationId && !organizationName.trim());
  return <div className="login-page"><section className="auth-card"><span className="eyebrow">Первый запуск</span><h1>Настроим пространство</h1>
    {!me.organization && <><p>Выберите организацию или создайте новую.</p><div className="form-grid"><label className="field">Организация<select value={organizationId} disabled={saving} onChange={(event) => setOrganizationId(event.target.value)}><option value="">Создать новую</option>{organizations.data?.map((organization) => <option value={organization.id} key={organization.id}>{organization.name}</option>)}</select></label>{!organizationId && <label className="field">Название организации<input value={organizationName} disabled={saving} onChange={(event) => setOrganizationName(event.target.value)} /></label>}</div></>}
    <p style={{ marginTop: 20 }}>Роли</p><div className="role-card-list">{roles.map((role) => <label className={selectedRoles.includes(role) ? 'role-card selected' : 'role-card'} key={role}><input className="sr-only" type="checkbox" disabled={saving} checked={selectedRoles.includes(role)} onChange={() => toggleRole(role)} /><span><strong>{roleName(role)}</strong><p>{descriptions[role]}</p></span>{selectedRoles.includes(role) && <b aria-hidden="true">✓</b>}</label>)}</div>
    {error && <p className="error" role="alert">{error}</p>}<button style={{ marginTop: 16 }} onClick={complete} disabled={cannotContinue}>{saving ? 'Сохраняем…' : 'Продолжить'}</button>
  </section></div>;
}
