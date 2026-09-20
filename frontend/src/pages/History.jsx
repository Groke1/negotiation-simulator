import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getSessionHistory, getUserId } from '../api';

export default function History() {
  const navigate = useNavigate();
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      const userId = getUserId();
      if (!userId) {
        setError('Нет данных пользователя. Войдите заново.');
        setLoading(false);
        return;
      }

      try {
        const data = await getSessionHistory(userId);
        setSessions(data.sessions || []);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) return <div className="page">Загрузка...</div>;

  return (
    <div className="page">
      <header>
        <h1>История тренировок</h1>
        <p>Всего сессий: {sessions.length}</p>
      </header>

      {error && <p className="error">{error}</p>}

      {!error && sessions.length === 0 && (
        <div className="card">
          <p className="muted">Вы ещё не проходили тренировки.</p>
          <button onClick={() => navigate('/home')}>Начать тренировку</button>
        </div>
      )}

      {sessions.map(s => (
        <div key={s.session_id} className="card session-card">
          <div className="session-header">
            <h3>{s.scenario_title}</h3>
            <span className="badge">
              {s.completed_at
                ? new Date(s.completed_at).toLocaleDateString('ru-RU')
                : '—'}
            </span>
          </div>
          <p className="scenario-desc">
            Режим: {s.mode} · Статус: {s.status}
          </p>
        </div>
      ))}

      <button onClick={() => navigate('/home')}>На главную</button>
    </div>
  );
}