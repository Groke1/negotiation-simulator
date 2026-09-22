import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Radar, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis,
  ResponsiveContainer
} from 'recharts';
import { getCuratorTeam, getUserProfile, getProfileHistory } from '../../api';

export default function CuratorHistory() {
  const navigate = useNavigate();
  const [team, setTeam] = useState([]);
  const [selectedUserId, setSelectedUserId] = useState('');
  const [profile, setProfile] = useState(null);
  const [history, setHistory] = useState(null);
  const [loading, setLoading] = useState(true);
  const [loadingProfile, setLoadingProfile] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      try {
        const teamData = await getCuratorTeam();
        setTeam(teamData);
        if (teamData.length > 0) {
          setSelectedUserId(teamData[0].user_id);
        }
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  useEffect(() => {
    async function loadProfile() {
      if (!selectedUserId) return;
      setLoadingProfile(true);
      setProfile(null);
      setHistory(null);
      try {
        const profileData = await getUserProfile(selectedUserId);
        setProfile(profileData);
        try {
          const historyData = await getProfileHistory(selectedUserId);
          setHistory(historyData);
        } catch {}
      } catch (err) {
        setError(err.message);
      } finally {
        setLoadingProfile(false);
      }
    }
    loadProfile();
  }, [selectedUserId]);

  if (loading) return <div className="page">Загрузка...</div>;

  const chartData = profile?.clusters?.length > 0
    ? profile.clusters.map(c => ({
        competency: c.name || c.cluster || c.cluster_id,
        level: c.level || c.score || 0
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
        <h1>История обучения</h1>
        <p>Результаты учеников по пройденным тренировкам.</p>
      </header>

      <div className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/curator')}>← Назад</button>
      </div>

      {error && <p className="error">{error}</p>}

      {team.length === 0 && (
        <div className="card">
          <p className="muted">У вас пока нет учеников в команде.</p>
        </div>
      )}

      {team.length > 0 && (
        <div className="card">
          <h2 className="section-title">Ученик</h2>
          <label className="field">
            <span>Выберите ученика</span>
            <select value={selectedUserId} onChange={e => setSelectedUserId(e.target.value)}>
              {team.map(m => (
                <option key={m.user_id} value={m.user_id}>
                  {m.name || m.username || m.email || m.user_id}
                </option>
              ))}
            </select>
          </label>
        </div>
      )}

      {selectedUserId && (
        <>
          <div className="card">
            <h2 className="section-title">Профиль компетенций</h2>

            {loadingProfile && <p className="muted">Загрузка...</p>}

            {!loadingProfile && !profile?.clusters?.length && (
              <p className="muted">
                У ученика пока нет результатов. Профиль строится по пройденным тренировкам.
              </p>
            )}

            {!loadingProfile && (
              <ResponsiveContainer width="100%" height={360}>
                <RadarChart data={chartData}>
                  <PolarGrid stroke="#e5e5e5" />
                  <PolarAngleAxis dataKey="competency" tick={{ fill: '#333', fontSize: 11 }} />
                  <PolarRadiusAxis domain={[0, 100]} tick={{ fill: '#999', fontSize: 11 }} />
                  <Radar dataKey="level" stroke="#111" fill="#111" fillOpacity={0.15} />
                </RadarChart>
              </ResponsiveContainer>
            )}
          </div>

          {history?.clusters?.length > 0 && (
            <div className="card">
              <h2 className="section-title">Динамика (до / после)</h2>
              {history.clusters.map((c, i) => (
                <div key={i} className="competency-item">
                  <span>{c.name || c.cluster}</span>
                  <span style={{ fontSize: 13, color: '#666' }}>
                    {c.before?.level ?? '—'} → {c.after?.level ?? '—'}
                  </span>
                </div>
              ))}
            </div>
          )}

          <div className="card">
            <h2 className="section-title">История тренировок</h2>
            {(!profile?.sessions || profile.sessions.length === 0) && (
              <p className="muted">Пока нет пройденных тренировок.</p>
            )}
            {profile?.sessions?.map((s, i) => (
              <div key={i} className="competency-item">
                <div style={{ flex: 1 }}>
                  <strong>{s.scenario_title || 'Тренировка'}</strong>
                  <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>
                    {s.status || '—'}
                    {s.completed_at && ` · ${new Date(s.completed_at).toLocaleDateString('ru-RU')}`}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}