export default function BlockPanel({ block, roles, allBlocks, onChange, onClose, onDelete }) {
  if (!block) return null;

  function update(field, value) {
    onChange({ ...block, [field]: value });
  }

  const isSystem = block.type === 'start' || block.type === 'end';
  const isNew = !!block.isNew;

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
            {isSystem ? 'Системный блок' : (isNew ? 'Новый блок' : 'Редактор блока')}
          </h3>
          <button onClick={onClose} style={{ padding: '6px 12px', fontSize: 14 }}>✕</button>
        </div>

        <label className="field">
          <span>Название</span>
          <input
            value={block.name || ''}
            onChange={e => update('name', e.target.value)}
            disabled={isSystem}
            autoFocus={isNew}
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
                rows={3}
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

            <div style={{ marginTop: 16 }}>
              <p className="muted" style={{ fontSize: 12, marginBottom: 8 }}>
                Переходы ({block.transitions?.length || 0})
              </p>

              {block.transitions?.map((tr, i) => (
                <div key={tr.id || i} style={{ marginBottom: 10, padding: 10, background: 'var(--paper)', borderRadius: 8, border: '1px solid var(--line)' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                    <input
                      value={tr.condition || ''}
                      onChange={e => {
                        const updated = [...block.transitions];
                        updated[i] = { ...tr, condition: e.target.value };
                        update('transitions', updated);
                      }}
                      placeholder="Условие перехода"
                      style={{ flex: 1, fontSize: 12, padding: '4px 8px', border: '1px solid var(--line)', borderRadius: 6, background: 'var(--paper-raised)', color: 'var(--ink)' }}
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

                  <select
                    value={tr.target_block_id || ''}
                    onChange={e => {
                      const updated = [...block.transitions];
                      updated[i] = { ...tr, target_block_id: e.target.value };
                      update('transitions', updated);
                    }}
                    style={{ width: '100%', fontSize: 12, padding: '6px 8px', border: '1px solid var(--line)', borderRadius: 6, background: 'var(--paper-raised)', color: 'var(--ink)' }}
                  >
                    <option value="">— куда ведёт —</option>
                    {(allBlocks || []).filter(b => b.id !== block.id).map(b => (
                      <option key={b.id} value={b.id}>{b.name}</option>
                    ))}
                  </select>
                </div>
              ))}

              <button
                onClick={() => {
                  const newTransition = {
                    id: 'tr-' + Date.now(),
                    condition: '',
                    target_block_id: '',
                    weight: 50,
                    forbidden: false
                  };
                  update('transitions', [...(block.transitions || []), newTransition]);
                }}
                style={{ fontSize: 12, padding: '6px 12px', marginTop: 6 }}
              >
                + Добавить переход
              </button>
            </div>

            <div style={{ marginTop: 24, display: 'flex', gap: 8 }}>
              {isNew ? (
                <button
                  onClick={onClose}
                  style={{
                    flex: 1,
                    background: 'var(--brand)',
                    color: 'var(--paper-raised)',
                    border: 'none',
                    padding: '12px',
                    fontWeight: 700,
                    borderRadius: 8,
                    fontSize: 14,
                    cursor: 'pointer',
                    fontFamily: 'inherit'
                  }}
                >
                  ✓ Сохранить блок
                </button>
              ) : (
                onDelete && (
                  <button
                    onClick={() => {
                      if (confirm('Удалить блок «' + block.name + '»? Все переходы к нему тоже удалятся.')) {
                        onDelete(block.id);
                        onClose();
                      }
                    }}
                    style={{
                      flex: 1,
                      background: 'var(--danger-soft)',
                      color: 'var(--danger)',
                      border: '1px solid var(--danger)',
                      padding: '12px',
                      fontWeight: 700,
                      borderRadius: 8,
                      fontSize: 14,
                      cursor: 'pointer',
                      fontFamily: 'inherit'
                    }}
                  >
                    🗑 Удалить блок
                  </button>
                )
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}