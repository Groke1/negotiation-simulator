import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getCuratorTeam, getCuratorAnalytics } from '../../api';

export default function Curator() {
  const navigate = useNavigate();
  const [team, setTeam] = useState([]);
  const [analytics, setAnalytics] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      try {
        const teamData = await getCuratorTeam();
        setTeam(teamData);
      } catch (err) {
        setError(err.message);
      }

      try {
        const analyticsData = await getCuratorAnalytics();
        setAnalytics(analyticsData);
      } catch {}

      setLoading(false);
    }
    load();
  }, []);

  if (loading) return <div className="page">Загрузка...</div>;

  return (
    <div className="page">
      <header>
        <h1>Дашборд куратора</h1>
        <p>Следите за успеваемостью учеников и назначайте тренировки.</p>
      </header>

      <div className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/curator/arena')}>Назначить арену</button>
        <button onClick={() => navigate('/curator/history')}>История обучения</button>
        <button onClick={() => navigate('/profile')}>Профиль</button>
      </div>

      {error && <p className="error">{error}</p>}

      <div className="card">
        <h2 className="section-title">Команда</h2>

        {team.length === 0 && (
          <p className="muted">У вас пока нет учеников в команде.</p>
        )}

        {team.map(member => (
          <div key={member.user_id} className="competency-item">
            <div style={{ flex: 1 }}>
              <strong>{member.name || member.username || 'Без имени'}</strong>
              <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>
                {member.email || '—'}
                {member.last_session_at && ` · последняя ${new Date(member.last_session_at).toLocaleDateString('ru-RU')}`}
              </p>
            </div>
            <button onClick={() => navigate(`/curator/learner/${member.user_id}`)}>
              Открыть
            </button>
          </div>
        ))}
      </div>

      {analytics && (
        <div className="card">
          <h2 className="section-title">Аналитика команды</h2>
          {analytics.clusters?.map((c, i) => (
            <div key={i} className="competency-item">
              <span>{c.cluster || c.name}</span>
              <span className="muted">{c.distribution || c.count || '—'}</span>
            </div>
          ))}
          {!analytics.clusters?.length && (
            <p className="muted">Аналитика пока пуста.</p>
          )}
        </div>
      )}

      {!analytics && (
        <div className="card">
          <h2 className="section-title">Аналитика команды</h2>
          <p className="muted">Аналитика пока недоступна.</p>
        </div>
      )}
    </div>
  );
}