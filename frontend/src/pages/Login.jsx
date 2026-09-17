import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login, register } from '../api';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [username, setUsername] = useState('');
  const [isRegister, setIsRegister] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    try {
      if (isRegister) await register(username, email, password);
      else await login(email, password);
      navigate('/home');
    } catch (err) {
      setError(err.message);
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

        {error && <p className="error">{error}</p>}

        <button type="submit">
          {isRegister ? 'Зарегистрироваться' : 'Войти'}
        </button>

        <button type="button" onClick={() => setIsRegister(!isRegister)}>
          {isRegister ? 'У меня есть аккаунт' : 'Создать аккаунт'}
        </button>
      </form>
    </div>
  );
}