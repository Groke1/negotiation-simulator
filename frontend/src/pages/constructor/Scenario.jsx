import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario, validateScenario, publishScenario, updateScenario } from '../../api';
import { mockScenario } from '../../data/mockScenario';
import { scenarioToFlow } from './tree/mapper';
import TreeEditor from './tree/TreeEditor';
import BlockPanel from './tree/BlockPanel';

export default function ScenarioView() {
  const { scenarioId } = useParams();
  const navigate = useNavigate();

  const [scenario, setScenario] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [tab, setTab] = useState('map');
  const [selectedBlockId, setSelectedBlockId] = useState(null);
  const [validation, setValidation] = useState(null);
  const [publishing, setPublishing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [lastSaved, setLastSaved] = useState(null);

  // Загрузка
  useEffect(() => {
    async function load() {
      try {
        const data = await getScenario(scenarioId);
        setScenario(data);
      } catch {
        // Пытаемся взять из localStorage
        const local = localStorage.getItem('scenario_' + scenarioId);
        if (local) {
          setScenario(JSON.parse(local));
        } else {
          setScenario({ ...mockScenario, scenario_id: scenarioId, isMock: true });
        }
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [scenarioId]);

  // Автосохранение в localStorage
  useEffect(() => {
    if (!scenario || loading) return;
    const timer = setTimeout(() => {
      localStorage.setItem('scenario_' + scenarioId, JSON.stringify(scenario));
      setLastSaved(new Date());
    }, 500);
    return () => clearTimeout(timer);
  }, [scenario, scenarioId, loading]);

  // Обновление блока
  function handleBlockChange(updatedBlock) {
    setScenario(prev => ({
      ...prev,
      blocks: prev.blocks.map(b => b.id === updatedBlock.id ? updatedBlock : b)
    }));
  }

  // Сохранение на бэк (для реальных сценариев)
  async function handleSave() {
    if (scenario.isMock || scenarioId.startsWith('sc-mock')) {
      // Мок — просто в localStorage
      localStorage.setItem('scenario_' + scenarioId, JSON.stringify(scenario));
      setLastSaved(new Date());
      return;
    }

    setSaving(true);
    setError('');
    try {
      await updateScenario(scenarioId, {
        title: scenario.title,
        description: scenario.description,
        difficulty: scenario.difficulty,
        tags: scenario.tags || [],
        supported_runtimes: scenario.supported_runtimes || [],
        supported_channels: scenario.supported_channels || [],
        target_cluster_ids: scenario.target_cluster_ids || [],
        roles: scenario.roles,
        blocks: scenario.blocks
      });
      setLastSaved(new Date());
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

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
  try {
    // Публикуем в localStorage
    const published = { ...scenario, status: 'published', publishedAt: new Date().toISOString() };
    localStorage.setItem('scenario_' + scenarioId, JSON.stringify(published));

    // Добавляем в общий список опубликованных
    const allPublished = JSON.parse(localStorage.getItem('published_scenarios') || '{}');
    allPublished[scenarioId] = published;
    localStorage.setItem('published_scenarios', JSON.stringify(allPublished));

    setScenario(published);
    alert('✅ Сценарий опубликован! Ученики увидят его в Обучении.');
  } catch (err) {
    setError(err.message);
  } finally {
    setPublishing(false);
  }
}

  if (loading) return <div className="page">Загрузка...</div>;
  if (error && !scenario) return <div className="page"><p className="error">{error}</p></div>;
  if (!scenario) return <div className="page">Сценарий не найден</div>;

  const flow = scenarioToFlow(scenario);
  const selectedBlock = selectedBlockId
    ? scenario.blocks.find(b => b.id === selectedBlockId)
    : null;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      <header className="chat-header">
        <button className="back-btn" onClick={() => navigate('/constructor')}>←</button>
        <div className="chat-header-info">
          <h2>{scenario.title}</h2>
          <p>
            {scenario.status} · {scenario.difficulty} · блоков: {scenario.blocks?.length || 0}
            {lastSaved && ` · сохранено ${lastSaved.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`}
          </p>
        </div>
        <div style={{ marginLeft: 'auto', display: 'flex', gap: 8 }}>
          <button onClick={handleSave} disabled={saving}>
            {saving ? 'Сохранение...' : 'Сохранить'}
          </button>
          <button onClick={handleValidate}>Проверить</button>
          {scenario.status === 'draft' && (
            <button onClick={handlePublish} disabled={publishing}>
              {publishing ? '...' : 'Опубликовать'}
            </button>
          )}
        </div>
      </header>

      {validation && (
        <div style={{ padding: '10px 20px', background: '#fafafa', borderBottom: '1px solid #eee' }}>
          <strong>Валидация:</strong> {validation.valid ? '✅ OK' : '❌ Есть ошибки'}
          {validation.errors_count > 0 && ` · ошибок: ${validation.errors_count}`}
          {validation.warnings_count > 0 && ` · предупреждений: ${validation.warnings_count}`}
        </div>
      )}

      <div style={{ display: 'flex', gap: 0, padding: '0 20px', borderBottom: '1px solid #eee' }}>
        {[
          { id: 'context', label: 'Контекст' },
          { id: 'map', label: 'Карта' }
        ].map(t => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            style={{
              border: 'none',
              background: 'transparent',
              padding: '12px 16px',
              fontSize: 14,
              fontWeight: 700,
              color: tab === t.id ? '#111' : '#888',
              borderBottom: tab === t.id ? '2px solid #111' : '2px solid transparent',
              cursor: 'pointer'
            }}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div style={{ flex: 1, display: 'flex', minHeight: 0 }}>
        {tab === 'context' && (
          <div style={{ flex: 1, padding: 24, overflowY: 'auto' }}>
            <div className="card">
              <h2>Описание</h2>
              <p>{scenario.description}</p>
            </div>
            {scenario.expanded_context && (
              <div className="card">
                <h2>Развёрнутый контекст</h2>
                <p>{scenario.expanded_context}</p>
              </div>
            )}
            <div className="card">
              <h2>Роли</h2>
              {scenario.roles?.map(r => (
                <div key={r.id} className="competency-item">
                  <span>{r.name}</span>
                  <span className="muted">{r.goal}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {tab === 'map' && (
          <div style={{ flex: 1, display: 'flex', minHeight: 0 }}>
            <div style={{ flex: 1 }}>
              <TreeEditor
                nodes={flow.nodes}
                edges={flow.edges}
                onSelectBlock={(block) => setSelectedBlockId(block.id)}
                onSelectEdge={() => {}}
              />
            </div>
            {selectedBlock && (
              <BlockPanel
                block={selectedBlock}
                roles={scenario.roles || []}
                onChange={handleBlockChange}
                onClose={() => setSelectedBlockId(null)}
              />
            )}
          </div>
        )}
      </div>
    </div>
  );
}