import { useMemo, useCallback } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
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
      stroke: e.data?.transition?.forbidden ? '#A6392E' : '#1A1D1B',
      strokeWidth: 2
    },
    labelStyle: { fontSize: 11, fill: '#565C58', fontWeight: 600 },
    labelBgStyle: { fill: '#FAFAF7', fillOpacity: 0.95 },
    labelBgPadding: [4, 2],
    labelBgBorderRadius: 4,
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: e.data?.transition?.forbidden ? '#A6392E' : '#1A1D1B'
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
        <Background color="#DCDAD1" gap={20} />
        <Controls />
        <MiniMap
          nodeColor={n => {
            const t = n.data?.type;
            if (t === 'start') return '#2F6F5E';
            if (t === 'end') return '#A6392E';
            return '#453B78';
          }}
        />
      </ReactFlow>

      {onAddBlock && (
        <button
          onClick={onAddBlock}
          style={{
            position: 'absolute',
            top: 16,
            left: 16,
            zIndex: 10,
            padding: '10px 16px',
            background: '#453B78',
            color: '#FAFAF7',
            border: 'none',
            borderRadius: 10,
            fontSize: 14,
            fontWeight: 700,
            cursor: 'pointer',
            fontFamily: 'inherit',
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)'
          }}
        >
          + Добавить блок
        </button>
      )}
    </div>
  );
}