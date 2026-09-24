import { useMemo, useCallback } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MarkerType
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import BlockNode from './BlockNode';

const nodeTypes = { blockNode: BlockNode };

export default function TreeEditor({
  nodes,
  edges,
  onNodeDragStop,
  onSelectBlock,
  onSelectEdge,
  onAddBlock,
  onAddTransition
}) {
  const styledEdges = useMemo(() => edges.map(e => ({
    ...e,
    animated: false,
    style: {
      stroke: e.data?.transition?.forbidden ? '#ef7c86' : '#a78bfa',
      strokeWidth: 2
    },
    labelStyle: { fontSize: 11, fill: '#aaa3b8', fontWeight: 600 },
    labelBgStyle: { fill: '#191722', fillOpacity: 0.95 },
    labelBgPadding: [4, 2],
    labelBgBorderRadius: 4,
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: e.data?.transition?.forbidden ? '#ef7c86' : '#a78bfa'
    }
  })), [edges]);

  const handleNodeClick = useCallback((_, node) => {
    if (onSelectBlock) onSelectBlock(node.data.block);
  }, [onSelectBlock]);

  const handleEdgeClick = useCallback((_, edge) => {
    if (onSelectEdge) onSelectEdge(edge.data);
  }, [onSelectEdge]);

  const handleConnect = useCallback((connection) => {
    if (onAddTransition && connection.source && connection.target) {
      onAddTransition(connection.source, connection.target);
    }
  }, [onAddTransition]);

  const handleNodeDragStop = useCallback((_, node) => {
    if (onNodeDragStop) onNodeDragStop(node.id, node.position);
  }, [onNodeDragStop]);

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <ReactFlow
        nodes={nodes}
        edges={styledEdges}
        nodeTypes={nodeTypes}
        onNodeClick={handleNodeClick}
        onEdgeClick={handleEdgeClick}
        onConnect={handleConnect}
        onNodeDragStop={handleNodeDragStop}
        fitView
      >
        <Background color="#343044" gap={20} />
        <Controls className="scenario-flow-controls" />
      </ReactFlow>

      {onAddBlock && (
        <button
          onClick={onAddBlock}
          className="canvas-add-button"
        >
          + Добавить блок
        </button>
      )}
    </div>
  );
}
