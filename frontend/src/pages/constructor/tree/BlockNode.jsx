import { Handle, Position } from '@xyflow/react';

const TYPE_COLORS = {
  start: '#0a7c2f',
  regular: '#111',
  end: '#dc2626'
};

const TYPE_LABELS = {
  start: 'Начало',
  regular: 'Блок',
  end: 'Финал'
};

export default function BlockNode({ data, selected }) {
  const block = data.block || {};
  const type = block.type || 'regular';
  const color = TYPE_COLORS[type] || '#111';

  return (
    <div style={{
      padding: '12px 16px',
      borderRadius: 12,
      background: 'white',
      border: `${selected ? 3 : 2}px solid ${color}`,
      minWidth: 200,
      maxWidth: 240,
      boxShadow: selected ? '0 8px 20px rgba(0,0,0,0.15)' : '0 2px 6px rgba(0,0,0,0.08)',
      fontFamily: 'system-ui, sans-serif',
      transition: 'box-shadow 0.15s'
    }}>
      <Handle type="target" position={Position.Top} />

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

      <div style={{ fontSize: 14, fontWeight: 600, color: '#0a0a0a', marginBottom: 4 }}>
        {data.name}
      </div>

      {data.description && (
        <div style={{ fontSize: 11, color: '#666', marginBottom: 6, lineHeight: 1.4 }}>
          {data.description.slice(0, 60)}{data.description.length > 60 ? '...' : ''}
        </div>
      )}

      {block.transitions?.length > 0 && (
        <div style={{ fontSize: 10, color: '#888', marginTop: 6 }}>
          Переходов: {block.transitions.length}
        </div>
      )}

      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}