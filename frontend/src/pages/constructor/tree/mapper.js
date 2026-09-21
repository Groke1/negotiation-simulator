function computeLevels(blocks) {
  const byId = Object.fromEntries(blocks.map(b => [b.id, b]));
  const start = blocks.find(b => b.type === 'start') || blocks[0];
  if (!start) return {};

  const levels = { [start.id]: 0 };
  const queue = [start.id];
  const visited = new Set([start.id]);

  while (queue.length) {
    const cur = queue.shift();
    const block = byId[cur];
    if (!block) continue;
    (block.transitions || []).forEach(tr => {
      if (!visited.has(tr.target_block_id) && byId[tr.target_block_id]) {
        visited.add(tr.target_block_id);
        levels[tr.target_block_id] = levels[cur] + 1;
        queue.push(tr.target_block_id);
      }
    });
  }

  let maxLevel = Math.max(0, ...Object.values(levels));
  blocks.forEach(b => {
    if (!(b.id in levels)) {
      maxLevel += 1;
      levels[b.id] = maxLevel;
    }
  });

  return levels;
}

export function scenarioToFlow(scenario) {
  if (!scenario || !scenario.blocks) return { nodes: [], edges: [] };

  const levels = computeLevels(scenario.blocks);
  const byLevel = {};
  scenario.blocks.forEach(b => {
    const lvl = levels[b.id] ?? 0;
    (byLevel[lvl] = byLevel[lvl] || []).push(b.id);
  });

  const nodes = [];
  Object.entries(byLevel).forEach(([lvl, ids]) => {
    ids.forEach((id, row) => {
      const block = scenario.blocks.find(b => b.id === id);
      nodes.push({
        id: block.id,
        type: 'blockNode',
        position: { x: Number(lvl) * 280, y: row * 150 },
        data: {
          name: block.name || 'Без названия',
          description: block.description || '',
          type: block.type || 'regular',
          block
        }
      });
    });
  });

  const edges = [];
  scenario.blocks.forEach(block => {
    (block.transitions || []).forEach(tr => {
      edges.push({
        id: tr.id,
        source: block.id,
        target: tr.target_block_id,
        label: tr.condition || '',
        data: { transition: tr, sourceBlockId: block.id }
      });
    });
  });

  return { nodes, edges };
}

export function flowToScenario(flow, baseScenario) {
  // обратное преобразование: собираем blocks из nodes + edges
  const blocks = flow.nodes.map(node => {
    const block = node.data.block;
    return {
      ...block,
      name: node.data.name,
      description: node.data.description
    };
  });

  return { ...baseScenario, blocks };
}