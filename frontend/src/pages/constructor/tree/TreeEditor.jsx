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

export default function TreeEditor({ nodes, edges, onSelectBlock, onSelectEdge }) {
  const styledEdges = useMemo(() => edges.map(e => ({
    ...e,
    animated: false,
    style: {
      stroke: e.data?.transition?.forbidden ? '#dc2626' : '#111',
      strokeWidth: 2
    },
    labelStyle: { fontSize: 11, fill: '#555', fontWeight: 600 },
    labelBgStyle: { fill: 'white', fillOpacity: 0.9 },
    labelBgPadding: [4, 2],
    labelBgBorderRadius: 4,
    markerEnd: {
      type: MarkerType.ArrowClosed,
      color: e.data?.transition?.forbidden ? '#dc2626' : '#111'
    }
  })), [edges]);

  const handleNodeClick = useCallback((_, node) => {
    if (onSelectBlock) onSelectBlock(node.data.block);
  }, [onSelectBlock]);

  const handleEdgeClick = useCallback((_, edge) => {
    if (onSelectEdge) onSelectEdge(edge.data);
  }, [onSelectEdge]);

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
        <MiniMap
          nodeColor={n => {
            const t = n.data?.type;
            if (t === 'start') return '#0a7c2f';
            if (t === 'end') return '#dc2626';
            return '#111';
          }}
        />
      </ReactFlow>
    </div>
  );
}