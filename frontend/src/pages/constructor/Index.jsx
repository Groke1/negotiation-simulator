import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getScenarios } from '../../api';

export default function Constructor() {
  const navigate = useNavigate();
  const [scenarios, setScenarios] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      try {
        const data = await getScenarios();
        setScenarios(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div className="page">
      <header>
        <h1>Конструктор сценариев</h1>
        <p>Здесь методист создаёт и публикует тренировки.</p>
      </header>

      <div className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/constructor/create')}>Создать сценарий</button>
        <button onClick={() => navigate('/profile')}>Профиль</button>
        <button onClick={() => navigate('/home')}>Обучение</button>
      </div>

      {loading && <p className="muted">Загрузка...</p>}
      {error && <p className="error">{error}</p>}

      {!loading && !error && scenarios.length === 0 && (
        <div className="card">
          <p className="muted">Библиотека пуста. Создайте первый сценарий.</p>
        </div>
      )}

      {scenarios.map(s => (
        <div key={s.scenario_id} className="card scenario-card">
          <div className="scenario-header">
            <h3>{s.title}</h3>
            <span className="badge">{s.difficulty}</span>
          </div>
          <p className="scenario-desc">Версия: {s.version}</p>
          <div className="scenario-meta">
            <span>Статус: <strong>{s.status}</strong></span>
            <span>Каналы: <strong>{s.channels?.join(', ') || '—'}</strong></span>
          </div>
          <button onClick={() => navigate(`/constructor/scenario/${s.scenario_id}`)}>
            Открыть
          </button>
        </div>
      ))}
    </div>
  );
}