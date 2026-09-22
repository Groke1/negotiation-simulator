export default function DialoguePanel({ node, block, onChange, onClose, isStart }) {
  if (!node) return null;

  function updateNode(field, value) {
    onChange({
      ...block,
      deterministic_dialogue: {
        ...block.deterministic_dialogue,
        nodes: block.deterministic_dialogue.nodes.map(n =>
          n.id === node.id ? { ...n, [field]: value } : n
        )
      }
    });
  }

  function updateIncomingMessage(text) {
    updateNode('incoming_message', {
      ...(node.incoming_message || {}),
      text,
      role_id: node.incoming_message?.role_id || block.lead_role_id || ''
    });
  }

  function updateReplyOption(optId, field, value) {
    updateNode('reply_options', (node.reply_options || []).map(o =>
      o.id === optId ? { ...o, [field]: value } : o
    ));
  }

  function addReplyOption() {
    updateNode('reply_options', [
      ...(node.reply_options || []),
      {
        id: 'ro-' + Date.now(),
        text: '',
        next_node_id: '',
        transition_id: '',
        semantic_action_code: ''
      }
    ]);
  }

  function removeReplyOption(optId) {
    updateNode('reply_options', (node.reply_options || []).filter(o => o.id !== optId));
  }

  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: 'rgba(0, 0, 0, 0.4)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 9999,
        fontFamily: 'inherit'
      }}
      onClick={onClose}
    >
      <div
        onClick={e => e.stopPropagation()}
        style={{
          width: 520,
          maxWidth: '90vw',
          maxHeight: '90vh',
          background: 'var(--paper-raised)',
          borderRadius: 14,
          padding: 24,
          overflowY: 'auto',
          boxShadow: '0 20px 60px rgba(0,0,0,0.3)'
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
          <h3 style={{ margin: 0, fontSize: 18 }}>
            {isStart ? 'Узел СТАРТ' : 'Редактор узла диалога'}
          </h3>
          <button onClick={onClose} style={{ padding: '6px 12px', fontSize: 14 }}>✕</button>
        </div>

        <label className="field">
          <span>Реплика партнёра</span>
          <textarea
            value={node.incoming_message?.text || ''}
            onChange={e => updateIncomingMessage(e.target.value)}
            placeholder="Что говорит партнёр..."
            rows={3}
            autoFocus
          />
        </label>

        <div style={{ marginTop: 16 }}>
          <p className="muted" style={{ fontSize: 12, marginBottom: 8 }}>
            Варианты ответа ({node.reply_options?.length || 0})
          </p>

          {(node.reply_options || []).map((opt, idx) => (
            <div key={opt.id} style={{
              padding: 10,
              background: 'var(--paper)',
              borderRadius: 8,
              marginBottom: 10,
              border: '1px solid var(--line)'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                <span style={{ fontSize: 11, fontWeight: 700, color: 'var(--ink-soft)' }}>
                  ВАРИАНТ {idx + 1}
                </span>
                <button
                  onClick={() => removeReplyOption(opt.id)}
                  style={{ padding: '2px 8px', fontSize: 11 }}
                >
                  ✕
                </button>
              </div>

              <textarea
                value={opt.text || ''}
                onChange={e => updateReplyOption(opt.id, 'text', e.target.value)}
                placeholder="Текст варианта..."
                rows={2}
                style={{ width: '100%', fontSize: 12, padding: '6px 8px', marginBottom: 8, background: 'var(--paper-raised)', color: 'var(--ink)', border: '1px solid var(--line)', borderRadius: 6 }}
              />

              <div style={{ fontSize: 11, color: 'var(--ink-soft)', marginBottom: 4 }}>
                Куда ведёт:
              </div>
              <select
                value={opt.transition_id || ''}
                onChange={e => {
                  updateReplyOption(opt.id, 'transition_id', e.target.value);
                  updateReplyOption(opt.id, 'next_node_id', '');
                }}
                style={{ fontSize: 11, width: '100%', padding: '4px 6px', background: 'var(--paper-raised)', color: 'var(--ink)', border: '1px solid var(--line)', borderRadius: 6 }}
              >
                <option value="">→ к переходу (в другой блок)</option>
                {(block.transitions || []).map(tr => (
                  <option key={tr.id} value={tr.id}>
                    {tr.condition || 'переход'}
                  </option>
                ))}
              </select>
            </div>
          ))}

          <button
            onClick={addReplyOption}
            style={{ fontSize: 12, padding: '6px 12px', marginTop: 4 }}
          >
            + Вариант ответа
          </button>
        </div>
      </div>
    </div>
  );
}