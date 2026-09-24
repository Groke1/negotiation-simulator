// Преобразование deterministic_dialogue блока в формат React Flow

export function dialogueToFlow(block) {
  if (!block?.deterministic_dialogue?.nodes?.length) {
    return { nodes: [], edges: [] };
  }

  const dialogue = block.deterministic_dialogue;
  const nodes = [];
  const edges = [];
  const positions = computeDialogueLayout(dialogue);

  dialogue.nodes.forEach(node => {
    nodes.push({
      id: node.id,
      type: 'dialogueNode',
      position: positions[node.id] || { x: 0, y: 0 },
      data: {
        node,
        isStart: node.id === dialogue.start_node_id
      }
    });

    (node.reply_options || []).forEach((opt) => {
      if (opt.next_node_id) {
        edges.push({
          id: `edge-${node.id}-${opt.id}`,
          source: node.id,
          target: opt.next_node_id,
          label: opt.text?.slice(0, 30) || '→',
          data: { reply: opt, sourceNodeId: node.id }
        });
      } else if (opt.transition_id) {
        // Ведёт к переходу — покажем как "выход"
        edges.push({
          id: `edge-${node.id}-${opt.id}`,
          source: node.id,
          target: `transition-${opt.transition_id}`,
          label: opt.text?.slice(0, 30) || '→ финал',
          data: { reply: opt, sourceNodeId: node.id, isTransition: true }
        });
      }
    });
  });

  return { nodes, edges };
}

// Простая вертикальная раскладка: старт сверху, потом по уровням
function computeDialogueLayout(dialogue) {
  const positions = {};
  const levels = {};
  const byId = Object.fromEntries(dialogue.nodes.map(n => [n.id, n]));

  // BFS от старта
  const queue = [[dialogue.start_node_id, 0]];
  const visited = new Set();

  while (queue.length) {
    const [nodeId, level] = queue.shift();
    if (visited.has(nodeId)) continue;
    visited.add(nodeId);
    levels[nodeId] = level;

    const node = byId[nodeId];
    if (!node) continue;

    (node.reply_options || []).forEach(opt => {
      if (opt.next_node_id && !visited.has(opt.next_node_id)) {
        queue.push([opt.next_node_id, level + 1]);
      }
    });
  }

  // Недостижимые — вниз
  let maxLevel = Math.max(0, ...Object.values(levels));
  dialogue.nodes.forEach(n => {
    if (!(n.id in levels)) {
      maxLevel += 1;
      levels[n.id] = maxLevel;
    }
  });

  // Разложить по уровням
  const byLevel = {};
  Object.entries(levels).forEach(([id, lvl]) => {
    (byLevel[lvl] = byLevel[lvl] || []).push(id);
  });

  Object.entries(byLevel).forEach(([lvl, ids]) => {
    ids.forEach((id, idx) => {
      positions[id] = {
        x: idx * 280,
        y: Number(lvl) * 200
      };
    });
  });

  return positions;
}
