export default function BlockPanel({ block, roles, onChange, onClose }) {
  if (!block) return null;

  function update(field, value) {
    onChange({ ...block, [field]: value });
  }

  const isSystem = block.type === 'start' || block.type === 'end';

  return (
    <div style={{
      width: 420,
      background: 'white',
      borderLeft: '1px solid #e0e0e0',
      padding: 20,
      overflowY: 'auto',
      height: '100%',
      fontFamily: 'system-ui, sans-serif'
    }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h3 style={{ margin: 0, fontSize: 16 }}>
          {isSystem ? 'Системный блок' : 'Редактор блока'}
        </h3>
        <button onClick={onClose} style={{ padding: '4px 10px', fontSize: 12 }}>✕</button>
      </div>

      <label className="field">
        <span>Название</span>
        <input
          value={block.name || ''}
          onChange={e => update('name', e.target.value)}
          disabled={isSystem}
        />
      </label>

      {!isSystem && (
        <>
          <label className="field">
            <span>Краткое описание</span>
            <textarea
              value={block.description || ''}
              onChange={e => update('description', e.target.value)}
              rows={2}
            />
          </label>

          <label className="field">
            <span>Развёрнутый контекст</span>
            <textarea
              value={block.expanded_context || ''}
              onChange={e => update('expanded_context', e.target.value)}
              rows={4}
            />
          </label>

          <label className="field">
            <span>Ведущий</span>
            <select
              value={block.lead_role_id || ''}
              onChange={e => update('lead_role_id', e.target.value || null)}
            >
              <option value="">— не задан —</option>
              {roles.map(r => (
                <option key={r.id} value={r.id}>{r.name}</option>
              ))}
            </select>
          </label>

          <label className="field">
            <span>Поведение</span>
            <textarea
              value={block.behavior || ''}
              onChange={e => update('behavior', e.target.value)}
              rows={2}
            />
          </label>

          {/* ==== ДЕЙСТВИЯ ==== */}
          <div style={{ marginTop: 16 }}>
            <p className="muted" style={{ fontSize: 12, marginBottom: 8 }}>
              Действия ({block.action_rules?.length || 0})
            </p>

            {block.action_rules?.map((rule, i) => (
              <div key={i} style={{ marginBottom: 10, padding: 10, background: '#fafafa', borderRadius: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                  <select
                    value={rule.semantic_action_code || ''}
                    onChange={e => {
                      const updated = [...block.action_rules];
                      updated[i] = { ...rule, semantic_action_code: e.target.value };
                      update('action_rules', updated);
                    }}
                    style={{ fontSize: 12, padding: '4px 8px', border: '1px solid #ddd', borderRadius: 6, flex: 1 }}
                  >
                    <option value="">— действие —</option>
                    <option value="ASK_QUESTION">Задать вопрос</option>
                    <option value="MAKE_PROPOSAL">Предложить</option>
                    <option value="MAKE_CONCESSION">Уступить</option>
                    <option value="REQUEST_RECIPROCITY">Запросить взаимность</option>
                    <option value="HOLD_BOUNDARY">Удержать границу</option>
                    <option value="PAUSE">Пауза</option>
                    <option value="SUMMARIZE">Резюмировать</option>
                  </select>
                  <button
                    onClick={() => {
                      const updated = block.action_rules.filter((_, idx) => idx !== i);
                      update('action_rules', updated);
                    }}
                    style={{ padding: '4px 8px', fontSize: 12, marginLeft: 8 }}
                  >
                    ✕
                  </button>
                </div>

                <textarea
                  value={rule.reaction || ''}
                  onChange={e => {
                    const updated = [...block.action_rules];
                    updated[i] = { ...rule, reaction: e.target.value };
                    update('action_rules', updated);
                  }}
                  placeholder="Что происходит в ответ..."
                  rows={2}
                  style={{ width: '100%', fontSize: 12, padding: '6px 8px', border: '1px solid #ddd', borderRadius: 6 }}
                />
              </div>
            ))}

            <button
              onClick={() => {
                const updated = [...(block.action_rules || []), { semantic_action_code: '', reaction: '' }];
                update('action_rules', updated);
              }}
              style={{ fontSize: 12, padding: '6px 12px', marginTop: 6 }}
            >
              + Добавить действие
            </button>
          </div>

          {/* ==== ОГРАНИЧЕНИЯ ==== */}
          <div style={{ marginTop: 16 }}>
            <p className="muted" style={{ fontSize: 12, marginBottom: 8 }}>
              Ограничения ({block.constraints?.length || 0})
            </p>

            {block.constraints?.map((c, i) => (
              <div key={i} style={{ display: 'flex', gap: 6, marginBottom: 6 }}>
                <input
                  value={c}
                  onChange={e => {
                    const updated = [...block.constraints];
                    updated[i] = e.target.value;
                    update('constraints', updated);
                  }}
                  style={{ flex: 1, fontSize: 12, padding: '6px 8px', border: '1px solid #ddd', borderRadius: 6 }}
                />
                <button
                  onClick={() => {
                    const updated = block.constraints.filter((_, idx) => idx !== i);
                    update('constraints', updated);
                  }}
                  style={{ padding: '4px 8px', fontSize: 12 }}
                >
                  ✕
                </button>
              </div>
            ))}

            <button
              onClick={() => {
                const updated = [...(block.constraints || []), ''];
                update('constraints', updated);
              }}
              style={{ fontSize: 12, padding: '6px 12px', marginTop: 6 }}
            >
              + Добавить ограничение
            </button>
          </div>

          {/* ==== ПЕРЕХОДЫ ==== */}
          <div style={{ marginTop: 16 }}>
            <p className="muted" style={{ fontSize: 12, marginBottom: 8 }}>
              Переходы ({block.transitions?.length || 0})
            </p>

            {block.transitions?.map((tr, i) => (
              <div key={tr.id || i} style={{ marginBottom: 10, padding: 10, background: '#fafafa', borderRadius: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                  <input
                    value={tr.condition || ''}
                    onChange={e => {
                      const updated = [...block.transitions];
                      updated[i] = { ...tr, condition: e.target.value };
                      update('transitions', updated);
                    }}
                    placeholder="Условие перехода"
                    style={{ flex: 1, fontSize: 12, padding: '4px 8px', border: '1px solid #ddd', borderRadius: 6 }}
                  />
                  <button
                    onClick={() => {
                      const updated = block.transitions.filter((_, idx) => idx !== i);
                      update('transitions', updated);
                    }}
                    style={{ padding: '4px 8px', fontSize: 12, marginLeft: 8 }}
                  >
                    ✕
                  </button>
                </div>
                <div style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                  <span style={{ fontSize: 11, color: '#666' }}>Вес:</span>
                  <input
                    type="number"
                    value={tr.weight ?? 50}
                    onChange={e => {
                      const updated = [...block.transitions];
                      updated[i] = { ...tr, weight: Number(e.target.value) };
                      update('transitions', updated);
                    }}
                    min={0}
                    max={100}
                    style={{ width: 60, fontSize: 12, padding: '4px 6px', border: '1px solid #ddd', borderRadius: 6 }}
                  />
                  <span style={{ fontSize: 11, color: '#666' }}>%</span>
                  <label style={{ fontSize: 11, marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 4 }}>
                    <input
                      type="checkbox"
                      checked={!!tr.forbidden}
                      onChange={e => {
                        const updated = [...block.transitions];
                        updated[i] = { ...tr, forbidden: e.target.checked };
                        update('transitions', updated);
                      }}
                    />
                    запрещён
                  </label>
                </div>
              </div>
            ))}
          </div>

          {/* ==== ДИАЛОГ ВНУТРИ БЛОКА ==== */}
          <div style={{ marginTop: 24, paddingTop: 16, borderTop: '2px solid #f0f0f0' }}>
            <p className="muted" style={{ fontSize: 13, marginBottom: 12, fontWeight: 700 }}>
              Диалог внутри блока ({block.deterministic_dialogue?.nodes?.length || 0} узлов)
            </p>

            {(!block.deterministic_dialogue || !block.deterministic_dialogue.nodes?.length) && (
              <p className="muted" style={{ fontSize: 12 }}>
                Нет диалога. Добавьте первый узел.
              </p>
            )}

            {block.deterministic_dialogue?.nodes?.map((node, nodeIdx) => (
              <div key={node.id} style={{
                border: '1px solid #e0e0e0',
                borderRadius: 8,
                padding: 12,
                marginBottom: 12,
                background: '#fafafa'
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <span style={{ fontSize: 11, fontWeight: 700, color: '#666' }}>
                    {nodeIdx === 0 ? 'СТАРТ' : `УЗЕЛ ${nodeIdx + 1}`}
                  </span>
                  <button
                    onClick={() => {
                      const updated = { ...block.deterministic_dialogue };
                      updated.nodes = updated.nodes.filter(n => n.id !== node.id);
                      update('deterministic_dialogue', updated);
                    }}
                    style={{ padding: '2px 8px', fontSize: 11 }}
                  >
                    ✕
                  </button>
                </div>

                {node.incoming_message && (
                  <div style={{ marginBottom: 8 }}>
                    <span style={{ fontSize: 11, color: '#666' }}>Реплика партнёра:</span>
                    <textarea
                      value={node.incoming_message.text || ''}
                      onChange={e => {
                        const updated = { ...block.deterministic_dialogue };
                        updated.nodes = updated.nodes.map(n =>
                          n.id === node.id
                            ? { ...n, incoming_message: { ...n.incoming_message, text: e.target.value } }
                            : n
                        );
                        update('deterministic_dialogue', updated);
                      }}
                      rows={2}
                      style={{ width: '100%', fontSize: 12, padding: '6px 8px', marginTop: 4 }}
                    />
                  </div>
                )}

                <div style={{ fontSize: 11, color: '#666', marginBottom: 4 }}>
                  Варианты ответа ({node.reply_options?.length || 0}):
                </div>

                {node.reply_options?.map(opt => (
                  <div key={opt.id} style={{
                    background: 'white',
                    border: '1px solid #ddd',
                    borderRadius: 6,
                    padding: 8,
                    marginBottom: 6
                  }}>
                    <textarea
                      value={opt.text || ''}
                      onChange={e => {
                        const updated = { ...block.deterministic_dialogue };
                        updated.nodes = updated.nodes.map(n =>
                          n.id === node.id
                            ? {
                                ...n,
                                reply_options: n.reply_options.map(o =>
                                  o.id === opt.id ? { ...o, text: e.target.value } : o
                                )
                              }
                            : n
                        );
                        update('deterministic_dialogue', updated);
                      }}
                      placeholder="Текст варианта..."
                      rows={2}
                      style={{ width: '100%', fontSize: 12, padding: '4px 6px' }}
                    />
                    <div style={{ display: 'flex', gap: 6, marginTop: 4 }}>
                      <select
                        value={opt.next_node_id || ''}
                        onChange={e => {
                          const updated = { ...block.deterministic_dialogue };
                          updated.nodes = updated.nodes.map(n =>
                            n.id === node.id
                              ? {
                                  ...n,
                                  reply_options: n.reply_options.map(o =>
                                    o.id === opt.id
                                      ? { ...o, next_node_id: e.target.value, transition_id: '' }
                                      : o
                                  )
                                }
                              : n
                          );
                          update('deterministic_dialogue', updated);
                        }}
                        style={{ fontSize: 11, flex: 1, padding: '4px 6px' }}
                      >
                        <option value="">→ к переходу</option>
                        {block.deterministic_dialogue.nodes
                          .filter(n => n.id !== node.id)
                          .map(n => (
                            <option key={n.id} value={n.id}>→ узел {n.id.slice(0, 6)}</option>
                          ))}
                      </select>
                      <select
                        value={opt.transition_id || ''}
                        onChange={e => {
                          const updated = { ...block.deterministic_dialogue };
                          updated.nodes = updated.nodes.map(n =>
                            n.id === node.id
                              ? {
                                  ...n,
                                  reply_options: n.reply_options.map(o =>
                                    o.id === opt.id
                                      ? { ...o, transition_id: e.target.value, next_node_id: '' }
                                      : o
                                  )
                                }
                              : n
                          );
                          update('deterministic_dialogue', updated);
                        }}
                        style={{ fontSize: 11, flex: 1, padding: '4px 6px' }}
                      >
                        <option value="">→ к переходу</option>
                        {block.transitions?.map(tr => (
                          <option key={tr.id} value={tr.id}>{tr.condition || 'переход'}</option>
                        ))}
                      </select>
                      <button
                        onClick={() => {
                          const updated = { ...block.deterministic_dialogue };
                          updated.nodes = updated.nodes.map(n =>
                            n.id === node.id
                              ? { ...n, reply_options: n.reply_options.filter(o => o.id !== opt.id) }
                              : n
                          );
                          update('deterministic_dialogue', updated);
                        }}
                        style={{ padding: '4px 8px', fontSize: 11 }}
                      >
                        ✕
                      </button>
                    </div>
                  </div>
                ))}

                <button
                  onClick={() => {
                    const updated = { ...block.deterministic_dialogue };
                    updated.nodes = updated.nodes.map(n =>
                      n.id === node.id
                        ? {
                            ...n,
                            reply_options: [
                              ...(n.reply_options || []),
                              {
                                id: 'ro-' + Date.now(),
                                text: '',
                                next_node_id: '',
                                transition_id: '',
                                semantic_action_code: ''
                              }
                            ]
                          }
                        : n
                    );
                    update('deterministic_dialogue', updated);
                  }}
                  style={{ fontSize: 11, padding: '4px 10px', marginTop: 4 }}
                >
                  + Вариант ответа
                </button>
              </div>
            ))}

            <button
              onClick={() => {
                const currentDialogue = block.deterministic_dialogue || { start_node_id: '', nodes: [] };
                const newNode = {
                  id: 'dn-' + Date.now(),
                  incoming_message: { text: '', role_id: block.lead_role_id || '' },
                  reply_options: []
                };
                const updated = {
                  ...currentDialogue,
                  start_node_id: currentDialogue.start_node_id || newNode.id,
                  nodes: [...currentDialogue.nodes, newNode]
                };
                update('deterministic_dialogue', updated);
              }}
              style={{ fontSize: 12, padding: '6px 12px', marginTop: 8 }}
            >
              + Добавить узел диалога
            </button>
          </div>
        </>
      )}

      {isSystem && (
        <p className="muted">
          {block.type === 'start' ? 'Точка входа в сценарий.' : 'Финальное состояние.'}
        </p>
      )}
    </div>
  );
}