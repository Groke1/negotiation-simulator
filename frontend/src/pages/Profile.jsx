import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  getMe,
  getRoles,
  switchRole,
  grantRole,
  getOrganizations,
  setOrganization,
  createOrganization,
  createOrganizationInternal
} from '../api';

export default function Profile() {
  const navigate = useNavigate();
  const [me, setMe] = useState(null);
  const [roles, setRoles] = useState(null);
  const [organizations, setOrganizations] = useState([]);
  const [selectedOrg, setSelectedOrg] = useState('');
  const [newOrgName, setNewOrgName] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [switching, setSwitching] = useState(false);
  const [orgSaving, setOrgSaving] = useState(false);
  const [orgCreating, setOrgCreating] = useState(false);

  async function reloadOrgs() {
    try {
      const orgsData = await getOrganizations();
      setOrganizations(orgsData);
    } catch {}
  }

  useEffect(() => {
    async function load() {
      try {
        const meData = await getMe();
        setMe(meData);
        const rolesData = await getRoles();
        setRoles(rolesData);

        await reloadOrgs();
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  async function handleSwitchRole(role) {
    setSwitching(true);
    setError('');
    try {
      await switchRole(role);
      if (role === 'methodist') navigate('/constructor');
      else if (role === 'curator') navigate('/curator');
      else navigate('/home');
    } catch (err) {
      setError(err.message);
    } finally {
      setSwitching(false);
    }
  }

  async function handleGrantRole(role) {
    setSwitching(true);
    setError('');
    try {
      await grantRole(me.id, role);
      const updatedRoles = await getRoles();
      setRoles(updatedRoles);
    } catch (err) {
      setError(err.message);
    } finally {
      setSwitching(false);
    }
  }

  async function handleSetOrg(e) {
    e.preventDefault();
    setError('');
    setSuccess('');
    setOrgSaving(true);
    try {
      await setOrganization(selectedOrg);
      const meData = await getMe();
      setMe(meData);
      setSelectedOrg('');
      setSuccess('Организация привязана');
    } catch (err) {
      setError(err.message);
    } finally {
      setOrgSaving(false);
    }
  }

  async function handleCreateOrg(e) {
    e.preventDefault();
    setError('');
    setSuccess('');
    setOrgCreating(true);
    try {
      await createOrganizationInternal(newOrgName);
      setNewOrgName('');
      await reloadOrgs();
      setSuccess('Организация создана. Выберите её ниже.');
    } catch (err) {
      try {
        await createOrganization(newOrgName);
        setNewOrgName('');
        await reloadOrgs();
        setSuccess('Организация создана. Выберите её ниже.');
      } catch (err2) {
        setError(err2.message);
      }
    } finally {
      setOrgCreating(false);
    }
  }

  function handleGoHome() {
    const active = roles?.active_role;
    if (active === 'methodist') navigate('/constructor');
    else if (active === 'curator') navigate('/curator');
    else navigate('/home');
  }

  if (loading) return <div className="page">Загрузка...</div>;
  if (error && !me) return <div className="page"><p className="error">{error}</p></div>;
  if (!me) return <div className="page">Нет данных</div>;

  return (
    <div className="page">
      <header>
        <h1>Профиль</h1>
        <p>{me.email}</p>
      </header>

      {error && <p className="error">{error}</p>}
      {success && <p style={{ color: '#0a7c2f', fontSize: 14, fontWeight: 600 }}>{success}</p>}

      <div className="card">
        <h2 className="section-title">Личные данные</h2>
        <div className="competency-item">
          <span>Имя пользователя</span>
          <strong>{me.username}</strong>
        </div>
        <div className="competency-item">
          <span>Email</span>
          <strong>{me.email}</strong>
        </div>
        <div className="competency-item">
          <span>Организация</span>
          <strong>{me.organization?.name || 'Не выбрана'}</strong>
        </div>
      </div>

      <div className="card">
        <h2 className="section-title">Организации</h2>

        {!me.organization && organizations.length > 0 && (
          <form onSubmit={handleSetOrg} style={{ marginBottom: 16 }}>
            <label className="field">
              <span>Выбрать организацию</span>
              <select value={selectedOrg} onChange={e => setSelectedOrg(e.target.value)} required>
                <option value="">— выберите —</option>
                {organizations.map(o => (
                  <option key={o.id} value={o.id}>{o.name}</option>
                ))}
              </select>
            </label>
            <button type="submit" disabled={orgSaving || !selectedOrg}>
              {orgSaving ? 'Сохранение...' : 'Привязать'}
            </button>
          </form>
        )}

        {me.organization && (
          <p className="muted" style={{ marginTop: 0 }}>
            Вы привязаны к организации: <strong>{me.organization.name}</strong>
          </p>
        )}

        <form onSubmit={handleCreateOrg} style={{ borderTop: '1px solid var(--line)', paddingTop: 16 }}>
          <label className="field">
            <span>Создать организацию</span>
            <input
              value={newOrgName}
              onChange={e => setNewOrgName(e.target.value)}
              placeholder="Например: Alabuga"
              required
              minLength={1}
              maxLength={200}
            />
          </label>
          <button type="submit" disabled={orgCreating || !newOrgName.trim()}>
            {orgCreating ? 'Создание...' : 'Создать организацию'}
          </button>
        </form>
      </div>

      <div className="card">
        <h2 className="section-title">Роли</h2>
        <p className="muted" style={{ marginTop: 0 }}>
          У вас может быть несколько ролей. Активная — выделена.
        </p>

        {roles?.items?.map(role => (
          <div key={role} className="role-row">
            <span className={`role-name ${roles.active_role === role ? 'active' : ''}`}>
              {formatRole(role)}
              {roles.active_role === role && ' · активна'}
            </span>
            {roles.active_role !== role && (
              <button onClick={() => handleSwitchRole(role)} disabled={switching}>
                Переключить
              </button>
            )}
          </div>
        ))}

        <div style={{ marginTop: 16, borderTop: '1px solid #f0f0f0', paddingTop: 16 }}>
          <p className="muted" style={{ margin: '0 0 8px' }}>
            Для демо — выдать роль:
          </p>
          {!roles?.items?.includes('methodist') && (
            <button onClick={() => handleGrantRole('methodist')} disabled={switching}>
              Стать методистом
            </button>
          )}
          {!roles?.items?.includes('curator') && (
            <button onClick={() => handleGrantRole('curator')} disabled={switching}>
              Стать куратором
            </button>
          )}
        </div>
      </div>

      <button onClick={handleGoHome}>На главную</button>
    </div>
  );
}

function formatRole(role) {
  const map = {
    learner: 'Обучающийся',
    methodist: 'Методист',
    curator: 'Куратор'
  };
  return map[role] || role;
}