import { Handle, Position } from '@xyflow/react';

const TYPE_COLORS = {
  start: '#2F6F5E',
  regular: '#453B78',
  end: '#A6392E'
};

const TYPE_LABELS = {
  start: 'Начало',
  regular: 'Блок',
  end: 'Финал'
};

export default function BlockNode({ data, selected }) {
  const block = data.block || {};
  const type = block.type || 'regular';
  const color = TYPE_COLORS[type] || '#1A1D1B';

  return (
    <div style={{
      padding: '12px 16px',
      borderRadius: 12,
      background: 'var(--surface-2)',
      border: `${selected ? 3 : 2}px solid ${color}`,
      minWidth: 200,
      maxWidth: 240,
      boxShadow: selected ? '0 8px 20px rgba(0,0,0,0.35)' : '0 2px 6px rgba(0,0,0,0.22)',
      fontFamily: 'inherit',
      transition: 'box-shadow 0.15s'
    }}>
      <Handle
        type="target"
        position={Position.Top}
        style={{ width: 10, height: 10, background: color }}
      />

      <div style={{
        fontSize: 10,
        fontWeight: 700,
        color,
        textTransform: 'uppercase',
        letterSpacing: 0.5,
        marginBottom: 6
      }}>
        {TYPE_LABELS[type]}
      </div>

      <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--text)', marginBottom: 4 }}>
        {data.name}
      </div>

      {data.description && (
        <div style={{ fontSize: 11, color: 'var(--muted)', marginBottom: 6, lineHeight: 1.4 }}>
          {data.description.slice(0, 60)}{data.description.length > 60 ? '...' : ''}
        </div>
      )}

      {block.transitions?.length > 0 && (
        <div style={{ fontSize: 10, color: 'var(--muted)', marginTop: 6 }}>
          Переходов: {block.transitions.length}
        </div>
      )}

      <Handle
        type="source"
        position={Position.Bottom}
        style={{ width: 10, height: 10, background: color }}
      />
    </div>
  );
}
