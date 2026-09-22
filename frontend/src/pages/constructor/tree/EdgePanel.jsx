export default function EdgePanel({ edgeData, scenario, onChange, onClose }) {
  if (!edgeData?.transition) return null;

  const { transition, sourceBlockId } = edgeData;
  const sourceBlock = scenario.blocks.find(b => b.id === sourceBlockId);

  function update(field, value) {
    if (!sourceBlock) return;
    onChange({
      ...sourceBlock,
      transitions: sourceBlock.transitions.map(tr =>
        tr.id === transition.id ? { ...tr, [field]: value } : tr
      )
    });
  }

  function remove() {
    if (!sourceBlock) return;
    onChange({
      ...sourceBlock,
      transitions: sourceBlock.transitions.filter(tr => tr.id !== transition.id)
    });
    onClose();
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
          width: 420,
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
          <h3 style={{ margin: 0, fontSize: 18 }}>Редактор перехода</h3>
          <button onClick={onClose} style={{ padding: '6px 12px', fontSize: 14 }}>✕</button>
        </div>

        <label className="field">
          <span>Условие перехода</span>
          <textarea
            value={transition.condition || ''}
            onChange={e => update('condition', e.target.value)}
            rows={3}
            placeholder="Например: Клиент согласился на обмен"
            autoFocus
          />
        </label>

        <label className="field">
          <span>Целевой блок</span>
          <select
            value={transition.target_block_id || ''}
            onChange={e => update('target_block_id', e.target.value)}
          >
            {scenario.blocks.map(b => (
              <option key={b.id} value={b.id}>{b.name}</option>
            ))}
          </select>
        </label>

        <label className="field">
          <span>Вес (0–100)</span>
          <input
            type="number"
            min={0}
            max={100}
            value={transition.weight ?? 50}
            onChange={e => update('weight', Number(e.target.value))}
          />
        </label>

        <label className="field">
          <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <input
              type="checkbox"
              checked={!!transition.forbidden}
              onChange={e => update('forbidden', e.target.checked)}
              style={{ width: 'auto' }}
            />
            Запрещён
          </span>
        </label>

        <button
          onClick={remove}
          style={{
            marginTop: 16,
            background: 'var(--danger-soft)',
            color: 'var(--danger)',
            border: '1px solid var(--danger)'
          }}
        >
          Удалить переход
        </button>
      </div>
    </div>
  );
}