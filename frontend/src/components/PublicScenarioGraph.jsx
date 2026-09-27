import { useMemo } from 'react';
import { Background, Controls, Handle, MarkerType, Position, ReactFlow } from '@xyflow/react';
import '@xyflow/react/dist/style.css';

const blockTypeLabel = (type) => ({
  start: 'Начало',
  end: 'Завершение',
  regular: 'Состояние переговоров',
  middle: 'Состояние переговоров'
}[type] || 'Состояние переговоров');

function PublicBlockNode({ data }) {
  return <article className={`public-scenario-node ${data.isCurrent ? 'current' : ''} ${data.type || ''}`}>
    <Handle type="target" position={Position.Left} isConnectable={false} />
    <span className="public-scenario-node-type">{data.isCurrent ? 'Текущее состояние' : blockTypeLabel(data.type)}</span>
    <strong>{data.name || 'Состояние без названия'}</strong>
    {data.description && <p>{data.description}</p>}
    <Handle type="source" position={Position.Right} isConnectable={false} />
  </article>;
}

const nodeTypes = { publicScenarioBlock: PublicBlockNode };

function layoutBlocks(blocks, currentBlockId) {
  const knownBlocks = new Map(blocks.map((block) => [block.id, block]));
  const start = blocks.find((block) => block.type === 'start') || blocks[0];
  const levels = {};

  if (start) {
    const queue = [start.id];
    levels[start.id] = 0;
    while (queue.length) {
      const blockId = queue.shift();
      const block = knownBlocks.get(blockId);
      (block?.transitions || []).forEach((transition) => {
        const targetId = transition.target_block_id;
        if (knownBlocks.has(targetId) && levels[targetId] === undefined) {
          levels[targetId] = levels[blockId] + 1;
          queue.push(targetId);
        }
      });
    }
  }

  let finalLevel = Math.max(0, ...Object.values(levels));
  blocks.forEach((block) => {
    if (levels[block.id] === undefined) {
      finalLevel += 1;
      levels[block.id] = finalLevel;
    }
  });

  const usedRows = {};
  const nodes = blocks.map((block) => {
    const level = levels[block.id];
    const row = usedRows[level] || 0;
    usedRows[level] = row + 1;
    return {
      id: block.id,
      type: 'publicScenarioBlock',
      position: { x: level * 300, y: row * 175 },
      data: { ...block, isCurrent: block.id === currentBlockId }
    };
  });

  const edges = blocks.flatMap((block) => (block.transitions || [])
    .filter((transition) => knownBlocks.has(transition.target_block_id))
    .map((transition, index) => ({
      id: transition.id || `${block.id}-${transition.target_block_id}-${index}`,
      source: block.id,
      target: transition.target_block_id,
      type: 'smoothstep',
      style: { stroke: '#a78bfa', strokeWidth: 1.8 },
      markerEnd: { type: MarkerType.ArrowClosed, color: '#a78bfa' }
    })));

  return { nodes, edges };
}

export default function PublicScenarioGraph({ blocks, currentBlockId }) {
  const { nodes, edges } = useMemo(
    () => layoutBlocks(blocks, currentBlockId),
    [blocks, currentBlockId]
  );

  return <div className="public-scenario-graph" aria-label="Граф блоков сценария">
    <ReactFlow
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      fitView
      fitViewOptions={{ padding: 0.25 }}
      nodesDraggable={false}
      nodesConnectable={false}
      elementsSelectable={false}
      panOnDrag
      zoomOnScroll
      zoomOnPinch
      proOptions={{ hideAttribution: true }}
    >
      <Background color="#383149" gap={20} size={1} />
      <Controls className="public-scenario-graph-controls" showInteractive={false} />
    </ReactFlow>
  </div>;
}
