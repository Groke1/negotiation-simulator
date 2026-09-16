import { useNavigate } from 'react-router-dom';
import { getHistory } from '../data/history';

export default function History() {
  const navigate = useNavigate();
  const sessions = getHistory();

  return (
    <div className="page">
      <header>
        <h1>История тренировок</h1>
        <p>Всего сессий: {sessions.length}</p>
      </header>

      {sessions.length === 0 && (
        <div className="card">
          <p className="muted">Вы ещё не проходили тренировки.</p>
          <button onClick={() => navigate('/')}>Начать</button>
        </div>
      )}

      {sessions.map(s => (
        <div key={s.id} className="card session-card">
          <div className="session-header">
            <h3>{s.scenarioTitle}</h3>
            <span className="badge">{s.date}</span>
          </div>
          <p className="scenario-desc">Роль: {s.role}</p>
          <button onClick={() => navigate(`/debrief/${s.scenarioId}/${s.outcome}`)}>
            Посмотреть разбор
          </button>
        </div>
      ))}

      <button className="back-home" onClick={() => navigate('/')}>На главную</button>
    </div>
  );
}