import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login, registerWithRole, getMe } from '../api';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [role, setRole] = useState('learner');
  const [isRegister, setIsRegister] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (isRegister) {
        await registerWithRole(username, email, password, role);
      } else {
        await login(email, password);
      }

      const me = await getMe();
      localStorage.setItem('user_id', me.id);

      if (me.active_role === 'methodist') navigate('/constructor');
      else if (me.active_role === 'curator') navigate('/curator');
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
          <label className="field">
            <span>Кто вы?</span>
            <select value={role} onChange={e => setRole(e.target.value)}>
              <option value="learner">Обучающийся</option>
              <option value="methodist">Методист</option>
              <option value="curator">Куратор</option>
            </select>
          </label>
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