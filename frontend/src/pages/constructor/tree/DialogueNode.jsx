import { Handle, Position } from '@xyflow/react';

export default function DialogueNode({ data, selected }) {
  const node = data.node || {};
  const isStart = data.isStart;

  return (
    <div style={{
      padding: '12px 16px',
      borderRadius: 12,
      background: 'white',
      border: `${selected ? 3 : 2}px solid ${isStart ? '#0a7c2f' : '#111'}`,
      minWidth: 220,
      maxWidth: 260,
      boxShadow: selected ? '0 8px 20px rgba(0,0,0,0.15)' : '0 2px 6px rgba(0,0,0,0.08)',
      fontFamily: 'system-ui, sans-serif',
      transition: 'box-shadow 0.15s'
    }}>
      <Handle type="target" position={Position.Top} />

      {isStart && (
        <div style={{
          fontSize: 10,
          fontWeight: 700,
          color: '#0a7c2f',
          textTransform: 'uppercase',
          letterSpacing: 0.5,
          marginBottom: 6
        }}>
          СТАРТ
        </div>
      )}

      {node.incoming_message?.text ? (
        <>
          <div style={{ fontSize: 10, color: '#888', marginBottom: 4 }}>
            Партнёр говорит:
          </div>
          <div style={{ fontSize: 13, color: '#111', marginBottom: 8, lineHeight: 1.4 }}>
            «{node.incoming_message.text.slice(0, 80)}
            {node.incoming_message.text.length > 80 ? '...' : ''}»
          </div>
        </>
      ) : (
        <div style={{ fontSize: 12, color: '#999', fontStyle: 'italic', marginBottom: 8 }}>
          Реплика партнёра не задана
        </div>
      )}

      <div style={{ fontSize: 10, color: '#666', marginTop: 8 }}>
        Вариантов ответа: {node.reply_options?.length || 0}
      </div>

      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}