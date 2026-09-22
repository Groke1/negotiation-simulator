import { useCallback } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  MarkerType
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import DialogueNode from './DialogueNode';

const nodeTypes = { dialogueNode: DialogueNode };

export default function DialogueTree({ nodes, edges, onSelectNode, onSelectReply }) {
  const styledEdges = edges.map(e => ({
    ...e,
    animated: false,
    style: {
      stroke: e.data?.isTransition ? '#dc2626' : '#111',
      strokeWidth: 2
    },
    labelStyle: { fontSize: 10, fill: '#555', fontWeight: 600 },
    labelBgStyle: { fill: 'white', fillOpacity: 0.9 },
    labelBgPadding: [4, 2],
    labelBgBorderRadius: 4,
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: e.data?.isTransition ? '#dc2626' : '#111'
    }
  }));

  const handleNodeClick = useCallback((_, node) => {
    if (onSelectNode) onSelectNode(node.data.node);
  }, [onSelectNode]);

  const handleEdgeClick = useCallback((_, edge) => {
    if (onSelectReply && edge.data?.reply) {
      onSelectReply(edge.data.reply, edge.data.sourceNodeId);
    }
  }, [onSelectReply]);

  if (nodes.length === 0) {
    return (
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100%',
        color: '#999',
        fontSize: 14,
        textAlign: 'center',
        padding: 40
      }}>
        <div>
          <p>Диалог внутри этого блока пуст.</p>
          <p style={{ fontSize: 12 }}>Добавь узлы во вкладке «Блоки» → панель блока → «Диалог внутри блока».</p>
        </div>
      </div>
    );
  }

  return (
    <div style={{ width: '100%', height: '100%' }}>
      <ReactFlow
        nodes={nodes}
        edges={styledEdges}
        nodeTypes={nodeTypes}
        onNodeClick={handleNodeClick}
        onEdgeClick={handleEdgeClick}
        fitView
      >
        <Background color="#e5e5e5" gap={20} />
        <Controls />
        <MiniMap nodeColor={n => n.data?.isStart ? '#0a7c2f' : '#111'} />
      </ReactFlow>
    </div>
  );
}