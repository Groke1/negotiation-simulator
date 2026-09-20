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

        try {
          const analyticsData = await getCuratorAnalytics();
          setAnalytics(analyticsData);
        } catch {
          // аналитика может не работать — не критично
        }
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
        <h1>Дашборд куратора</h1>
        <p>Следите за успеваемостью учеников и назначайте тренировки.</p>
      </header>

      <section className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/curator/arena')}>Назначить арену</button>
        <button onClick={() => navigate('/profile')}>Профиль</button>
        <button onClick={() => navigate('/home')}>Обучение</button>
      </section>

      {error && <p className="error">{error}</p>}

      <div className="card">
        <h2 className="section-title">Команда</h2>

        {team.length === 0 && !error && (
          <p className="muted">У вас пока нет учеников в команде.</p>
        )}

        {team.map(member => (
          <div key={member.user_id} className="competency-item">
            <div style={{ flex: 1 }}>
              <strong>{member.name}</strong>
              <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>
                Слабое место: {member.weakest_cluster || '—'}
              </p>
              <p className="muted" style={{ margin: '2px 0 0', fontSize: 12 }}>
                Последняя сессия: {member.last_session_at
                  ? new Date(member.last_session_at).toLocaleDateString('ru-RU')
                  : 'не проходил'}
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

          {analytics.most_common_weak_cluster && (
            <p className="muted">
              Самое частое слабое место: <strong>{analytics.most_common_weak_cluster}</strong>
            </p>
          )}

          {analytics.clusters?.map(c => (
            <div key={c.cluster} className="competency-item">
              <span>{c.cluster}</span>
              <span style={{ fontSize: 13, color: '#666' }}>
                0–1: {c.distribution.level_0_1} · 2–3: {c.distribution.level_2_3} · 4: {c.distribution.level_4}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}