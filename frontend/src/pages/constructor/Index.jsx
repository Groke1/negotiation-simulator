import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getScenarios, archiveScenario } from '../../api';

export default function Constructor() {
  const navigate = useNavigate();
  const [scenarios, setScenarios] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');

  async function load() {
    setLoading(true);
    setError('');
    try {
      const data = await getScenarios();
      setScenarios(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function handleArchive(scenarioId, title) {
    if (!window.confirm(`Архивировать сценарий «${title}»?`)) return;
    try {
      await archiveScenario(scenarioId);
      setScenarios(prev => prev.filter(s => s.scenario_id !== scenarioId));
    } catch (err) {
      alert(err.message);
    }
  }

  const filtered = scenarios.filter(s => {
    if (statusFilter === 'all') return true;
    return s.status === statusFilter;
  });

  const draftCount = scenarios.filter(s => s.status === 'draft').length;
  const publishedCount = scenarios.filter(s => s.status === 'published').length;
  const archivedCount = scenarios.filter(s => s.status === 'archived').length;

  return (
    <div className="page">
      <header>
        <h1>Конструктор сценариев</h1>
        <p>Создавайте и публикуйте тренировки.</p>
      </header>

      <div className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/constructor/create')}>+ Создать сценарий</button>
        <button onClick={() => navigate('/profile')}>Профиль</button>
      </div>

      <div style={{ display: 'flex', gap: 6, marginBottom: 16, flexWrap: 'wrap' }}>
        {[
          { id: 'all', label: 'Все', count: scenarios.length },
          { id: 'draft', label: 'Черновики', count: draftCount },
          { id: 'published', label: 'Опубликованные', count: publishedCount },
          { id: 'archived', label: 'Архив', count: archivedCount }
        ].map(f => (
          <button
            key={f.id}
            onClick={() => setStatusFilter(f.id)}
            style={{
              border: `1.5px solid ${statusFilter === f.id ? 'var(--brand)' : 'var(--line)'}`,
              background: statusFilter === f.id ? 'var(--brand)' : 'transparent',
              color: statusFilter === f.id ? 'var(--paper-raised)' : 'var(--ink-soft)',
              borderRadius: 20,
              padding: '6px 14px',
              fontSize: 13,
              fontWeight: 700,
              cursor: 'pointer',
              fontFamily: 'inherit'
            }}
          >
            {f.label} · {f.count}
          </button>
        ))}
      </div>

      {loading && <p className="muted">Загрузка...</p>}
      {error && <p className="error">{error}</p>}

      {!loading && !error && filtered.length === 0 && (
        <div className="card">
          <p className="muted">Нет сценариев в этой категории. Создайте первый.</p>
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
          </div>
          <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
            <button onClick={() => navigate(`/constructor/scenario/${s.scenario_id}`)}>
              Открыть
            </button>
            {s.status !== 'archived' && (
              <button
                onClick={() => handleArchive(s.scenario_id, s.title)}
                style={{ background: 'transparent', color: '#c1121f', border: '1.5px solid #c1121f' }}
              >
                Архивировать
              </button>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}