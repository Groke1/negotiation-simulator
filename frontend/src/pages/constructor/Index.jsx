import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getScenarios } from '../../api';
import { mockScenario } from '../../data/mockScenario';

export default function Constructor() {
  const navigate = useNavigate();
  const [scenarios, setScenarios] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  useEffect(() => {
    async function load() {
      try {
        const data = await getScenarios();
        // Добавляем мок-сценарий в список, чтобы демо работало даже с пустым бэком
        setScenarios([
          { ...mockScenario, isMock: true },
          ...data
        ]);
      } catch (err) {
        // Если бэк 404 — показываем только мок
        setScenarios([{ ...mockScenario, isMock: true }]);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const filtered = scenarios.filter(s => {
    if (statusFilter === 'all') return true;
    return s.status === statusFilter;
  });

  return (
    <div className="page">
      <header>
        <h1>Конструктор сценариев</h1>
        <p>Создавайте и публикуйте тренировки.</p>
      </header>

      <div className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/constructor/create')}>+ Создать сценарий</button>
        <button onClick={() => navigate('/profile')}>Профиль</button>
        <button onClick={() => navigate('/home')}>Обучение</button>
      </div>

      <div style={{ display: 'flex', gap: 6, marginBottom: 16 }}>
        {[
          { id: 'all', label: 'Все', count: scenarios.length },
          { id: 'draft', label: 'Черновики', count: scenarios.filter(s => s.status === 'draft').length },
          { id: 'published', label: 'Опубликованные', count: scenarios.filter(s => s.status === 'published').length }
        ].map(f => (
          <button
            key={f.id}
            onClick={() => setStatusFilter(f.id)}
            style={{
              border: `1.5px solid ${statusFilter === f.id ? '#111' : '#ddd'}`,
              background: statusFilter === f.id ? '#111' : 'transparent',
              color: statusFilter === f.id ? '#fff' : '#666',
              borderRadius: 20,
              padding: '6px 14px',
              fontSize: 13,
              fontWeight: 700,
              cursor: 'pointer'
            }}
          >
            {f.label} · {f.count}
          </button>
        ))}
      </div>

      {loading && <p className="muted">Загрузка...</p>}
      {error && <p className="error">{error}</p>}

      {!loading && filtered.length === 0 && (
        <div className="card">
          <p className="muted">Нет сценариев. Создайте первый.</p>
        </div>
      )}

      {filtered.map(s => (
        <div key={s.scenario_id} className="card scenario-card">
          <div className="scenario-header">
            <h3>{s.title}</h3>
            <span className="badge">{s.difficulty}</span>
          </div>
          <p className="scenario-desc">{s.description}</p>
          <div className="scenario-meta">
            <span>Статус: <strong>{s.status}</strong></span>
            <span>Теги: <strong>{s.tags?.join(', ') || '—'}</strong></span>
            {s.isMock && <span className="muted">демо-сценарий</span>}
          </div>
          <button onClick={() => navigate(`/constructor/scenario/${s.scenario_id}`)}>
            Открыть
          </button>
        </div>
      ))}
    </div>
  );
}