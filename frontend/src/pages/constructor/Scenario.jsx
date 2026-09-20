import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario, validateScenario, publishScenario } from '../../api';

export default function ScenarioView() {
  const { scenarioId } = useParams();
  const navigate = useNavigate();
  const [scenario, setScenario] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [publishing, setPublishing] = useState(false);
  const [validation, setValidation] = useState(null);

  useEffect(() => {
    async function load() {
      try {
        const data = await getScenario(scenarioId);
        setScenario(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [scenarioId]);

  async function handleValidate() {
    try {
      const result = await validateScenario(scenarioId);
      setValidation(result);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handlePublish() {
    setPublishing(true);
    setError('');
    try {
      await publishScenario(scenarioId);
      const updated = await getScenario(scenarioId);
      setScenario(updated);
    } catch (err) {
      setError(err.message);
    } finally {
      setPublishing(false);
    }
  }

  if (loading) return <div className="page">Загрузка...</div>;
  if (error) return <div className="page"><p className="error">{error}</p></div>;
  if (!scenario) return <div className="page">Сценарий не найден</div>;

  return (
    <div className="page">
      <header>
        <h1>{scenario.title}</h1>
        <p>Версия {scenario.version} · {scenario.difficulty} · {scenario.status}</p>
      </header>

      <div className="card">
        <h2 className="section-title">Контекст</h2>
        <p>{scenario.context?.description}</p>
      </div>

      <div className="card">
        <h2 className="section-title">Роли</h2>
        {scenario.roles?.map(r => (
          <div key={r.id} className="competency-item">
            <span>{r.id}</span>
            <strong>{r.label}</strong>
          </div>
        ))}
      </div>

      <div className="card">
        <h2 className="section-title">Узлы сценария</h2>
        {scenario.nodes?.map(n => (
          <div
            key={n.id}
            className="competency-item"
            style={{ flexDirection: 'column', alignItems: 'flex-start', gap: 8 }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
              <strong>{n.id}</strong>
              <span className="muted">{n.type}</span>
            </div>
            <p style={{ margin: 0, fontSize: 14 }}>{n.text}</p>
            <p className="muted" style={{ margin: 0, fontSize: 12 }}>
              Действий: {n.actions?.length || 0}
            </p>
          </div>
        ))}
      </div>

      {validation && (
        <div className="card">
          <h2 className="section-title">Результат проверки</h2>
          <p>Валиден: <strong>{validation.valid ? 'Да' : 'Нет'}</strong></p>
          {validation.graph?.dead_ends?.length > 0 && (
            <p className="error">Тупики: {validation.graph.dead_ends.join(', ')}</p>
          )}
        </div>
      )}

      <div className="nav">
        <button onClick={handleValidate}>Проверить</button>
        {scenario.status === 'draft' && (
          <button onClick={handlePublish} disabled={publishing}>
            {publishing ? 'Публикация...' : 'Опубликовать'}
          </button>
        )}
        <button onClick={() => navigate('/constructor')}>Назад</button>
      </div>
    </div>
  );
}