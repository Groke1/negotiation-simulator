import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { login, registerWithRolesAndOrg, getMe, getRoles, getOrganizations } from '../api';

const AVAILABLE_ROLES = [
  { id: 'learner', label: 'Обучающийся' },
  { id: 'methodist', label: 'Методист' },
  { id: 'curator', label: 'Куратор' }
];

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [selectedRoles, setSelectedRoles] = useState(['learner']);
  const [organizationId, setOrganizationId] = useState('');
  const [organizations, setOrganizations] = useState([]);
  const [isRegister, setIsRegister] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    if (isRegister) {
      getOrganizations()
        .then(setOrganizations)
        .catch(() => setOrganizations([]));
    }
  }, [isRegister]);

  function toggleRole(roleId) {
    setSelectedRoles(prev => {
      if (prev.includes(roleId)) {
        if (prev.length === 1) return prev;
        return prev.filter(r => r !== roleId);
      }
      return [...prev, roleId];
    });
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (isRegister) {
        if (selectedRoles.length === 0) {
          throw new Error('Выберите хотя бы одну роль');
        }
        await registerWithRolesAndOrg(
          username,
          email,
          password,
          selectedRoles,
          organizationId || null
        );
      } else {
        await login(email, password);
      }

      const me = await getMe();
      localStorage.setItem('user_id', me.id);

      // Определяем активную роль через getRoles()
      let activeRole = 'learner';
      try {
        const rolesData = await getRoles();
        activeRole = rolesData.active_role;
      } catch {}

      if (activeRole === 'methodist') navigate('/constructor');
      else if (activeRole === 'curator') navigate('/curator');
      else navigate('/home');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page">
      <header>
        <h1>{isRegister ? 'Регистрация' : 'Вход'}</h1>
        <p>Войдите, чтобы начать тренировки.</p>
      </header>

      <form onSubmit={handleSubmit} className="card">
        {isRegister && (
          <label className="field">
            <span>Имя пользователя</span>
            <input
              value={username}
              onChange={e => setUsername(e.target.value)}
              required
              minLength={3}
            />
          </label>
        )}

        <label className="field">
          <span>Email</span>
          <input
            type="email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            required
          />
        </label>

        <label className="field">
          <span>Пароль</span>
          <input
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            required
            minLength={isRegister ? 8 : 1}
          />
        </label>

        {isRegister && (
          <>
            <div className="field">
              <span>Роли</span>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginTop: 6 }}>
                {AVAILABLE_ROLES.map(role => {
                  const active = selectedRoles.includes(role.id);
                  return (
                    <button
                      key={role.id}
                      type="button"
                      onClick={() => toggleRole(role.id)}
                      style={{
                        border: `1.5px solid ${active ? '#111' : '#ddd'}`,
                        background: active ? '#111' : 'transparent',
                        color: active ? '#fff' : '#666',
                        borderRadius: 20,
                        padding: '6px 14px',
                        fontSize: 13,
                        fontWeight: 700,
                        cursor: 'pointer',
                        fontFamily: 'inherit'
                      }}
                    >
                      {active ? '✓ ' : ''}{role.label}
                    </button>
                  );
                })}
              </div>
              <p className="muted" style={{ fontSize: 12, marginTop: 4 }}>
                Можно выбрать несколько. Одна из них будет активной.
              </p>
            </div>

            <label className="field">
              <span>Организация</span>
              <select
                value={organizationId}
                onChange={e => setOrganizationId(e.target.value)}
              >
                <option value="">— не выбрана —</option>
                {organizations.map(org => (
                  <option key={org.id} value={org.id}>{org.name}</option>
                ))}
              </select>
              {organizations.length === 0 && (
                <p className="muted" style={{ fontSize: 12, marginTop: 4 }}>
                  Список организаций пуст. Можно оставить «не выбрана».
                </p>
              )}
            </label>
          </>
        )}

        {error && <p className="error">{error}</p>}

        <button type="submit" disabled={loading}>
          {loading ? 'Загрузка...' : (isRegister ? 'Зарегистрироваться' : 'Войти')}
        </button>

        <button type="button" onClick={() => setIsRegister(!isRegister)} disabled={loading}>
          {isRegister ? 'У меня есть аккаунт' : 'Создать аккаунт'}
        </button>
      </form>
    </div>
  );
}