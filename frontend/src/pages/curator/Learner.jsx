import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Radar, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis,
  ResponsiveContainer
} from 'recharts';
import { getProfile, getSessionHistory, getProfileHistory } from '../../api';

export default function Learner() {
  const { userId } = useParams();
  const navigate = useNavigate();
  const [profile, setProfile] = useState(null);
  const [sessions, setSessions] = useState([]);
  const [history, setHistory] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      try {
        const profileData = await getProfile(userId);
        setProfile(profileData);

        try {
          const sessionsData = await getSessionHistory(userId);
          setSessions(sessionsData.sessions || []);
        } catch {
          // история может не работать
        }

        try {
          const historyData = await getProfileHistory(userId);
          setHistory(historyData);
        } catch {
          // динамика может не работать
        }
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [userId]);

  if (loading) return <div className="page">Загрузка...</div>;
  if (error) return <div className="page"><p className="error">{error}</p></div>;

  const chartData = profile?.clusters?.length > 0
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
        <h1>Ученик</h1>
        <p>ID: {userId}</p>
      </header>

      <div className="card">
        <h2 className="section-title">Профиль компетенций</h2>
        <ResponsiveContainer width="100%" height={360}>
          <RadarChart data={chartData}>
            <PolarGrid stroke="#e5e5e5" />
            <PolarAngleAxis dataKey="competency" tick={{ fill: '#333', fontSize: 11 }} />
            <PolarRadiusAxis domain={[0, 4]} tick={{ fill: '#999', fontSize: 11 }} />
            <Radar dataKey="level" stroke="#111" fill="#111" fillOpacity={0.15} />
          </RadarChart>
        </ResponsiveContainer>
      </div>

      {history && history.clusters?.length > 0 && (
        <div className="card">
          <h2 className="section-title">Динамика (до/после)</h2>
          {history.clusters.map((c, i) => (
            <div key={i} className="competency-item">
              <span>{c.cluster}</span>
              <span style={{ fontSize: 13, color: '#666' }}>
                {c.before?.level ?? '—'} → {c.after?.level ?? '—'}
              </span>
            </div>
          ))}
        </div>
      )}

      <div className="card">
        <h2 className="section-title">История тренировок</h2>

        {sessions.length === 0 && (
          <p className="muted">Ученик ещё не проходил тренировки.</p>
        )}

        {sessions.map(s => (
          <div key={s.session_id} className="competency-item">
            <div>
              <strong>{s.scenario_title}</strong>
              <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>
                {s.mode} · {s.status} · {s.completed_at
                  ? new Date(s.completed_at).toLocaleDateString('ru-RU')
                  : '—'}
              </p>
            </div>
          </div>
        ))}
      </div>

      <button onClick={() => navigate('/curator')}>Назад к команде</button>
    </div>
  );
}