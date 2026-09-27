import { useMemo, useState } from 'react';
import { Background, Controls, Handle, MarkerType, Position, ReactFlow } from '@xyflow/react';
import '@xyflow/react/dist/style.css';

const blockTypeLabel = (type) => ({
  start: 'Начало',
  end: 'Завершение',
  regular: 'Состояние переговоров',
  middle: 'Состояние переговоров'
}[type] || 'Состояние переговоров');

function PublicBlockNode({ data }) {
  return <article className={`public-scenario-node ${data.isCurrent ? 'current' : ''} ${data.isSelected ? 'selected' : ''} ${data.isRelated ? 'related' : ''} ${data.type || ''}`}>
    <Handle type="target" position={Position.Left} isConnectable={false} />
    <span className="public-scenario-node-type">{data.isCurrent ? 'Текущее состояние' : blockTypeLabel(data.type)}</span>
    <strong>{data.name || 'Состояние без названия'}</strong>
    {data.description && <p>{data.description}</p>}
    <Handle type="source" position={Position.Right} isConnectable={false} />
  </article>;
}

const nodeTypes = { publicScenarioBlock: PublicBlockNode };

function layoutBlocks(blocks, currentBlockId, selectedId) {
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

  const relatedBlockIds = new Set();
  if (selectedId) {
    relatedBlockIds.add(selectedId);
    blocks.forEach((block) => {
      if (block.id === selectedId) (block.transitions || []).forEach((transition) => relatedBlockIds.add(transition.target_block_id));
      if ((block.transitions || []).some((transition) => transition.target_block_id === selectedId)) relatedBlockIds.add(block.id);
    });
  }

  const usedRows = {};
  const nodes = blocks.map((block) => {
    const level = levels[block.id];
    const row = usedRows[level] || 0;
    usedRows[level] = row + 1;
    return {
      id: block.id,
      type: 'publicScenarioBlock',
      position: { x: level * 300, y: row * 175 },
      data: { ...block, isCurrent: block.id === currentBlockId, isSelected: block.id === selectedId, isRelated: selectedId && relatedBlockIds.has(block.id) && block.id !== selectedId }
    };
  });

  const edges = blocks.flatMap((block) => (block.transitions || [])
    .filter((transition) => knownBlocks.has(transition.target_block_id))
    .map((transition, index) => {
      const isRelated = Boolean(selectedId) && (block.id === selectedId || transition.target_block_id === selectedId);
      const color = isRelated ? '#72d6b1' : selectedId ? '#574d6e' : '#a78bfa';
      return {
        id: transition.id || `${block.id}-${transition.target_block_id}-${index}`,
        source: block.id,
        target: transition.target_block_id,
        type: 'smoothstep',
        animated: isRelated,
        className: isRelated ? 'public-scenario-edge-active' : 'public-scenario-edge',
        style: { stroke: color, strokeWidth: isRelated ? 3 : 1.8, opacity: selectedId && !isRelated ? 0.45 : 1 },
        markerEnd: { type: MarkerType.ArrowClosed, color }
      };
    }));

  return { nodes, edges };
}

export default function PublicScenarioGraph({ blocks, currentBlockId }) {
  const [selectedId, setSelectedId] = useState(null);
  const { nodes, edges } = useMemo(
    () => layoutBlocks(blocks, currentBlockId, selectedId),
    [blocks, currentBlockId, selectedId]
  );
  const selectedBlock = blocks.find((block) => block.id === selectedId);

  return <div className="public-scenario-graph-wrap">
    <div className="public-scenario-graph" aria-label="Граф блоков сценария">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        onNodeClick={(_, node) => setSelectedId(node.id === selectedId ? null : node.id)}
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
    </div>
    {selectedBlock && <section className="public-block-details" aria-live="polite">
      <div><span className="eyebrow">{blockTypeLabel(selectedBlock.type)}</span><h4>{selectedBlock.name || 'Состояние без названия'}</h4></div>
      <button className="secondary" type="button" onClick={() => setSelectedId(null)}>Скрыть</button>
      <p>{selectedBlock.description || 'Описание этого состояния не добавлено.'}</p>
    </section>}
  </div>;
}
