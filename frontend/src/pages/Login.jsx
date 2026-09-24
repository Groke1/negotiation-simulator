import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { login, register } from '../api';
import { useAuth } from '../app/AuthContext';

export default function Login({ register: isRegistration = false }) {
  const navigate = useNavigate();
  const location = useLocation();
  const { signInComplete } = useAuth();
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function submit(event) {
    event.preventDefault();
    const values = new FormData(event.currentTarget);
    const username = String(values.get('username') || '').trim();
    const email = String(values.get('email') || '').trim();
    const password = String(values.get('password') || '');

    setError('');
    setSubmitting(true);
    try {
      if (isRegistration) await register(username, email, password);
      else await login(email, password);
      await signInComplete();
      navigate(location.state?.from?.pathname || '/', { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="login-page">
      <section className="auth-card">
        <span className="eyebrow">Арена переговоров</span>
        <h1>{isRegistration ? 'Создать аккаунт' : 'С возвращением'}</h1>
        <p>{isRegistration ? 'Начните практиковать переговоры в реальных сценариях.' : 'Войдите, чтобы продолжить обучение.'}</p>
        <form className="form-grid" onSubmit={submit} style={{ marginTop: 24 }}>
          {isRegistration && (
            <label className="field">
              Имя
              <input name="username" autoComplete="name" required minLength="2" disabled={submitting} />
            </label>
          )}
          <label className="field">
            Email
            <input name="email" type="email" autoComplete="email" required disabled={submitting} />
          </label>
          <label className="field">
            Пароль
            <input name="password" type="password" autoComplete={isRegistration ? 'new-password' : 'current-password'} required minLength="6" disabled={submitting} />
          </label>
          {error && <p className="error" role="alert">{error}</p>}
          <button type="submit" disabled={submitting}>{submitting ? 'Подождите…' : isRegistration ? 'Зарегистрироваться' : 'Войти'}</button>
        </form>
        <footer>
          {isRegistration ? <>Уже есть аккаунт? <Link to="/auth/login">Войти</Link></> : <>Нет аккаунта? <Link to="/auth/register">Зарегистрироваться</Link></>}
        </footer>
      </section>
    </div>
  );
}
