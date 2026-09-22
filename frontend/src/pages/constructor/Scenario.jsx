import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario, validateScenario, publishScenario, updateScenario } from '../../api';
import { mockScenario } from '../../data/mockScenario';
import { getScenario as getLocalScenario } from '../../data/scenario';
import { scenarioToFlow } from './tree/mapper';
import { dialogueToFlow } from './tree/dialogueMapper';
import TreeEditor from './tree/TreeEditor';
import DialogueTree from './tree/DialogueTree';
import BlockPanel from './tree/BlockPanel';
import DialoguePanel from './tree/DialoguePanel';
import EdgePanel from './tree/EdgePanel';

export default function ScenarioView() {
  const { scenarioId } = useParams();
  const navigate = useNavigate();

  const [scenario, setScenario] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [tab, setTab] = useState('map');
  const [mapMode, setMapMode] = useState('blocks');
  const [selectedBlockId, setSelectedBlockId] = useState(null);
  const [selectedDialogueNodeId, setSelectedDialogueNodeId] = useState(null);
  const [selectedEdge, setSelectedEdge] = useState(null);
  const [validation, setValidation] = useState(null);
  const [publishing, setPublishing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [lastSaved, setLastSaved] = useState(null);

  useEffect(() => {
    async function load() {
      try {
        const data = await getScenario(scenarioId);
        setScenario(data);
      } catch {
        // 1. localStorage
        const local = localStorage.getItem('scenario_' + scenarioId);
        if (local) {
          setScenario(JSON.parse(local));
          setLoading(false);
          return;
        }

        // 2. Хардкодные сценарии из scenario.js
        const localScenario = getLocalScenario(scenarioId);
        if (localScenario) {
          setScenario({ ...localScenario, isMock: true });
          setLoading(false);
          return;
        }

        // 3. mockScenario
        setScenario({ ...mockScenario, scenario_id: scenarioId, isMock: true });
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [scenarioId]);

  useEffect(() => {
    if (!scenario || loading) return;
    const timer = setTimeout(() => {
      localStorage.setItem('scenario_' + scenarioId, JSON.stringify(scenario));
      setLastSaved(new Date());
    }, 500);
    return () => clearTimeout(timer);
  }, [scenario, scenarioId, loading]);

  const handleCloseEdge = useCallback(() => setSelectedEdge(null), []);
  const handleCloseDialogueNode = useCallback(() => setSelectedDialogueNodeId(null), []);

  const handleCloseBlock = useCallback(() => {
    if (selectedBlockId) {
      setScenario(prev => ({
        ...prev,
        blocks: prev.blocks.map(b =>
          b.id === selectedBlockId ? { ...b, isNew: false } : b
        )
      }));
    }
    setSelectedBlockId(null);
  }, [selectedBlockId]);

  const handleBlockChange = useCallback((updatedBlock) => {
    setScenario(prev => ({
      ...prev,
      blocks: prev.blocks.map(b => b.id === updatedBlock.id ? updatedBlock : b)
    }));
  }, []);

  const handleNodeDragStop = useCallback((nodeId, position) => {
    setScenario(prev => ({
      ...prev,
      blocks: prev.blocks.map(b =>
        b.id === nodeId ? { ...b, position } : b
      )
    }));
  }, []);

  const handleDeleteBlock = useCallback((blockId) => {
    setScenario(prev => ({
      ...prev,
      blocks: prev.blocks
        .filter(b => b.id !== blockId)
        .map(b => ({
          ...b,
          transitions: (b.transitions || []).filter(t => t.target_block_id !== blockId)
        }))
    }));
    setSelectedBlockId(null);
  }, []);

  function handleAddBlock() {
    const newId = 'block-' + Date.now();
    const newBlock = {
      id: newId,
      type: 'regular',
      name: 'Новый блок',
      description: '',
      expanded_context: '',
      lead_role_id: null,
      lead_probability: 50,
      role_goals: [],
      behavior: '',
      action_rules: [],
      constraints: [],
      transitions: [],
      position: { x: 100 + Math.random() * 400, y: 100 + Math.random() * 300 },
      isNew: true,
      deterministic_dialogue: {
        start_node_id: '',
        nodes: []
      }
    };

    setScenario(prev => ({
      ...prev,
      blocks: [...prev.blocks, newBlock]
    }));

    setSelectedBlockId(newId);
    setSelectedEdge(null);
    setSelectedDialogueNodeId(null);
  }

  function handleAddTransition(sourceBlockId, targetBlockId) {
    setScenario(prev => {
      const sourceBlock = prev.blocks.find(b => b.id === sourceBlockId);
      if (!sourceBlock) return prev;

      const exists = sourceBlock.transitions?.some(t => t.target_block_id === targetBlockId);
      if (exists) return prev;

      const newTransition = {
        id: 'tr-' + Date.now(),
        condition: '',
        target_block_id: targetBlockId,
        weight: 50,
        forbidden: false
      };

      return {
        ...prev,
        blocks: prev.blocks.map(b =>
          b.id === sourceBlockId
            ? { ...b, transitions: [...(b.transitions || []), newTransition] }
            : b
        )
      };
    });
  }

  async function handleSave() {
    if (scenario.isMock || scenarioId.startsWith('sc-mock')) {
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
      const published = {
        ...scenario,
        status: 'published',
        publishedAt: new Date().toISOString()
      };
      localStorage.setItem('scenario_' + scenarioId, JSON.stringify(published));

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

  const blockFlow = scenarioToFlow(scenario);

  const activeBlockId = selectedBlockId || scenario.blocks.find(x => x.type === 'regular')?.id;
  const selectedBlock = scenario.blocks.find(b => b.id === activeBlockId);

  const dialogueFlow = selectedBlock ? dialogueToFlow(selectedBlock) : { nodes: [], edges: [] };

  const selectedDialogueNode = selectedBlock?.deterministic_dialogue?.nodes?.find(
    n => n.id === selectedDialogueNodeId
  );

  const isStartNode = selectedBlock?.deterministic_dialogue?.start_node_id === selectedDialogueNodeId;

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
        <div style={{ padding: '10px 20px', background: 'var(--paper)', borderBottom: '1px solid var(--line)' }}>
          <strong>Валидация:</strong> {validation.valid ? '✅ OK' : '❌ Есть ошибки'}
          {validation.errors_count > 0 && ` · ошибок: ${validation.errors_count}`}
          {validation.warnings_count > 0 && ` · предупреждений: ${validation.warnings_count}`}
        </div>
      )}

      <div style={{ display: 'flex', gap: 0, padding: '0 20px', borderBottom: '1px solid var(--line)', background: 'var(--paper-raised)' }}>
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
              color: tab === t.id ? 'var(--ink)' : 'var(--ink-soft)',
              borderBottom: tab === t.id ? '2px solid var(--brand)' : '2px solid transparent',
              cursor: 'pointer',
              fontFamily: 'inherit'
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
          <div style={{ flex: 1, display: 'flex', minHeight: 0, flexDirection: 'column' }}>
            <div style={{
              padding: '10px 20px',
              background: 'var(--paper)',
              borderBottom: '1px solid var(--line)',
              display: 'flex',
              gap: 8,
              alignItems: 'center'
            }}>
              <span style={{ fontSize: 12, color: 'var(--ink-soft)', marginRight: 8 }}>Режим:</span>
              <button
                onClick={() => {
                  setMapMode('blocks');
                  setSelectedDialogueNodeId(null);
                }}
                style={{
                  padding: '6px 14px',
                  fontSize: 13,
                  fontWeight: 700,
                  borderRadius: 20,
                  border: `1.5px solid ${mapMode === 'blocks' ? 'var(--brand)' : 'var(--line)'}`,
                  background: mapMode === 'blocks' ? 'var(--brand)' : 'transparent',
                  color: mapMode === 'blocks' ? 'var(--paper-raised)' : 'var(--ink-soft)',
                  cursor: 'pointer',
                  fontFamily: 'inherit'
                }}
              >
                Блоки
              </button>
              <button
                onClick={() => {
                  setMapMode('dialogue');
                  setSelectedBlockId(null);
                }}
                style={{
                  padding: '6px 14px',
                  fontSize: 13,
                  fontWeight: 700,
                  borderRadius: 20,
                  border: `1.5px solid ${mapMode === 'dialogue' ? 'var(--brand)' : 'var(--line)'}`,
                  background: mapMode === 'dialogue' ? 'var(--brand)' : 'transparent',
                  color: mapMode === 'dialogue' ? 'var(--paper-raised)' : 'var(--ink-soft)',
                  cursor: 'pointer',
                  fontFamily: 'inherit'
                }}
              >
                Диалог внутри блока
              </button>

              {mapMode === 'dialogue' && (
                <select
                  value={activeBlockId || ''}
                  onChange={e => {
                    setSelectedBlockId(e.target.value);
                    setSelectedDialogueNodeId(null);
                  }}
                  style={{
                    marginLeft: 'auto',
                    padding: '6px 12px',
                    fontSize: 13,
                    border: '1px solid var(--line)',
                    borderRadius: 8,
                    background: 'var(--paper-raised)',
                    color: 'var(--ink)',
                    fontFamily: 'inherit'
                  }}
                >
                  {scenario.blocks.map(b => (
                    <option key={b.id} value={b.id}>
                      {b.name} ({b.deterministic_dialogue?.nodes?.length || 0} узлов)
                    </option>
                  ))}
                </select>
              )}
            </div>

            <div style={{ flex: 1, display: 'flex', minHeight: 0 }}>
              <div style={{ flex: 1 }}>
                {mapMode === 'blocks' ? (
                  <TreeEditor
                    nodes={blockFlow.nodes}
                    edges={blockFlow.edges}
                    onNodeDragStop={handleNodeDragStop}
                    onSelectBlock={(block) => {
                      setSelectedBlockId(block.id);
                      setSelectedDialogueNodeId(null);
                      setSelectedEdge(null);
                    }}
                    onSelectEdge={(data) => {
                      setSelectedEdge(data);
                      setSelectedBlockId(null);
                    }}
                    onAddBlock={handleAddBlock}
                    onAddTransition={handleAddTransition}
                  />
                ) : (
                  <DialogueTree
                    nodes={dialogueFlow.nodes}
                    edges={dialogueFlow.edges}
                    onSelectNode={(node) => setSelectedDialogueNodeId(node.id)}
                    onSelectReply={() => {}}
                  />
                )}
              </div>

              {mapMode === 'blocks' && selectedBlockId && (
                <BlockPanel
                  block={scenario.blocks.find(b => b.id === selectedBlockId) || null}
                  roles={scenario.roles || []}
                  allBlocks={scenario.blocks}
                  onChange={handleBlockChange}
                  onClose={handleCloseBlock}
                  onDelete={handleDeleteBlock}
                />
              )}

              {mapMode === 'blocks' && selectedEdge && (
                <EdgePanel
                  edgeData={selectedEdge}
                  scenario={scenario}
                  onChange={handleBlockChange}
                  onClose={handleCloseEdge}
                />
              )}

              {mapMode === 'dialogue' && selectedDialogueNode && selectedBlock && (
                <DialoguePanel
                  node={selectedDialogueNode}
                  selectedBlock={selectedBlock}
                  block={selectedBlock}
                  isStart={isStartNode}
                  onChange={handleBlockChange}
                  onClose={handleCloseDialogueNode}
                />
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}