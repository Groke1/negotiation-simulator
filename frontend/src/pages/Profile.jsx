import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Radar, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis,
  ResponsiveContainer
} from 'recharts';
import {
  getMe,
  getRoles,
  switchRole,
  getProfile,
  grantRole,
  getOrganizations,
  setOrganization
} from '../api';

export default function Profile() {
  const navigate = useNavigate();
  const [me, setMe] = useState(null);
  const [roles, setRoles] = useState(null);
  const [profile, setProfile] = useState(null);
  const [organizations, setOrganizations] = useState([]);
  const [selectedOrg, setSelectedOrg] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [switching, setSwitching] = useState(false);
  const [orgSaving, setOrgSaving] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const meData = await getMe();
        setMe(meData);
        const rolesData = await getRoles();
        setRoles(rolesData);

        try {
          const profileData = await getProfile(meData.id);
          setProfile(profileData);
        } catch {}

        try {
          const orgsData = await getOrganizations();
          setOrganizations(orgsData);
        } catch {}
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
    setOrgSaving(true);
    try {
      await setOrganization(selectedOrg);
      const meData = await getMe();
      setMe(meData);
      setSelectedOrg('');
    } catch (err) {
      setError(err.message);
    } finally {
      setOrgSaving(false);
    }
  }

  if (loading) return <div className="page">Загрузка...</div>;
  if (error && !me) return <div className="page"><p className="error">{error}</p></div>;
  if (!me) return <div className="page">Нет данных</div>;

  const hasProfileData = profile?.clusters && profile.clusters.length > 0;

  const chartData = hasProfileData
    ? profile.clusters.map(c => ({
        competency: c.cluster,
        level: c.level
      }))
    : [
        { competency: 'Цели и границы', level: 0 },
        { competency: 'Диагностика', level: 0 },
        { competency: 'Создание ценности', level: 0 },
        { competency: 'Аргументация', level: 0 },
        { competency: 'Торг и обмен', level: 0 },
        { competency: 'Конфликт', level: 0 }
      ];

  return (
    <div className="page">
      <header>
        <h1>Профиль</h1>
        <p>{me.email}</p>
      </header>

      {error && <p className="error">{error}</p>}

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

      {!me.organization && organizations.length > 0 && (
        <div className="card">
          <h2 className="section-title">Выбрать организацию</h2>
          <form onSubmit={handleSetOrg}>
            <label className="field">
              <span>Организация</span>
              <select value={selectedOrg} onChange={e => setSelectedOrg(e.target.value)} required>
                <option value="">— выберите —</option>
                {organizations.map(o => (
                  <option key={o.id} value={o.id}>{o.name}</option>
                ))}
              </select>
            </label>
            <button type="submit" disabled={orgSaving}>
              {orgSaving ? 'Сохранение...' : 'Привязать'}
            </button>
          </form>
        </div>
      )}

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

      <div className="card">
        <h2 className="section-title">Профиль компетенций</h2>

        {!hasProfileData && (
          <p className="muted" style={{ marginTop: 0, marginBottom: 12 }}>
            Профиль пока пуст. Пройдите тренировку, чтобы увидеть свои компетенции.
          </p>
        )}

        <ResponsiveContainer width="100%" height={360}>
          <RadarChart data={chartData}>
            <PolarGrid stroke="#e5e5e5" />
            <PolarAngleAxis dataKey="competency" tick={{ fill: '#333', fontSize: 11 }} />
            <PolarRadiusAxis domain={[0, 4]} tick={{ fill: '#999', fontSize: 11 }} />
            <Radar dataKey="level" stroke="#111" fill="#111" fillOpacity={0.15} />
          </RadarChart>
        </ResponsiveContainer>
      </div>

      <button onClick={() => navigate('/home')}>На главную</button>
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