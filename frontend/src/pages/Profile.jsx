import { useNavigate } from 'react-router-dom';
import {
  Radar, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis,
  ResponsiveContainer, Legend
} from 'recharts';
import { getHistory } from '../data/history';
import {
  competencies, computeLevels, computeOverallProfile, chartColors
} from '../data/profile';

export default function Profile() {
  const navigate = useNavigate();
  const sessions = getHistory();

  // Общий профиль (среднее по всем сессиям)
  const overall = computeOverallProfile(sessions);

  // Данные для общего радара
  const overallData = competencies.map(c => ({
    competency: c.name,
    level: overall[c.id]
  }));

  // Сессии для мульти-радара: максимум 4
  const recentSessions = sessions.slice(0, 4);

  // Собираем данные для мульти-радара
  const multiData = competencies.map((c) => {
    const row = { competency: c.name };
    recentSessions.forEach((s, i) => {
      const levels = computeLevels(s.patterns || []);
      row['s' + i] = levels[c.id];
    });
    return row;
  });

  return (
    <div className="page">
      <header>
        <h1>Профиль компетенций</h1>
        <p>Уровень по 6 кластерам: от 0 (не проявляется) до 4 (адаптивный).</p>
      </header>

      {sessions.length === 0 ? (
        <div className="card">
          <p className="muted">Пройдите хотя бы одну тренировку, чтобы увидеть профиль.</p>
          <button onClick={() => navigate('/')}>Начать</button>
        </div>
      ) : (
        <>
          <div className="card">
            <h2 className="section-title">Общий профиль</h2>
            <p className="muted" style={{ marginTop: 0, marginBottom: 12 }}>
              Среднее по {sessions.length} {sessions.length === 1 ? 'сессии' : 'сессиям'}.
            </p>
            <ResponsiveContainer width="100%" height={360}>
              <RadarChart data={overallData} outerRadius="70%">
                <PolarGrid stroke="#e5e5e5" />
                <PolarAngleAxis dataKey="competency" tick={{ fill: '#333', fontSize: 11 }} />
                <PolarRadiusAxis domain={[0, 4]} tick={{ fill: '#999', fontSize: 11 }} />
                <Radar dataKey="level" stroke="#111" fill="#111" fillOpacity={0.15} />
              </RadarChart>
            </ResponsiveContainer>
          </div>

          {recentSessions.length > 0 && (
            <div className="card">
              <h2 className="section-title">Последние {recentSessions.length} сессий</h2>
              <p className="muted" style={{ marginTop: 0, marginBottom: 12 }}>
                Каждый цвет — одна тренировка. Хранятся последние 4.
              </p>
              <ResponsiveContainer width="100%" height={420}>
                <RadarChart data={multiData} outerRadius="65%">
                  <PolarGrid stroke="#e5e5e5" />
                  <PolarAngleAxis dataKey="competency" tick={{ fill: '#333', fontSize: 11 }} />
                  <PolarRadiusAxis domain={[0, 4]} tick={{ fill: '#999', fontSize: 11 }} />
                  <Legend wrapperStyle={{ fontSize: 11 }} />
                  {recentSessions.map((s, i) => (
                    <Radar
                      key={i}
                      name={s.scenarioTitle}
                      dataKey={'s' + i}
                      stroke={chartColors[i]}
                      fill={chartColors[i]}
                      fillOpacity={0.08}
                    />
                  ))}
                </RadarChart>
              </ResponsiveContainer>
            </div>
          )}
        </>
      )}

      <div className="card">
        <h2 className="section-title">Расшифровка уровней</h2>
        <ul className="clean-list">
          <li>0 — не проявляется</li>
          <li>1 — начальный</li>
          <li>2 — рабочий</li>
          <li>3 — уверенный</li>
          <li>4 — адаптивный</li>
        </ul>
      </div>

      <button onClick={() => navigate('/')}>На главную</button>
    </div>
  );
}