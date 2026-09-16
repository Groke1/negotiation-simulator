import { scenario } from './data/scenario';

const USE_MOCK = true; // Поменяешь на false, когда бэк будет готов
const BASE_URL = 'http://localhost:8080';

const MOCK_PROFILE = {
  competencies: [
    { id: 'goals', name: 'Цели и границы', level: 2 },
    { id: 'diagnostics', name: 'Диагностика', level: 3 },
    { id: 'value', name: 'Создание ценности', level: 1 },
    { id: 'argumentation', name: 'Аргументация', level: 2 },
    { id: 'trade', name: 'Торг и обмен', level: 2 },
    { id: 'conflict', name: 'Управление конфликтом', level: 1 }
  ]
};

export async function getProfile() {
  if (USE_MOCK) return MOCK_PROFILE;
  const res = await fetch(`${BASE_URL}/profile`);
  return res.json();
}

export async function getNode(nodeId) {
  if (USE_MOCK) {
    return scenario.nodes.find(n => n.id === nodeId);
  }
  const res = await fetch(`${BASE_URL}/sessions/${nodeId}`);
  return res.json();
}

export async function sendAction(sessionId, actionId) {
  if (USE_MOCK) {
    return { next: actionId };
  }
  const res = await fetch(`${BASE_URL}/sessions/${sessionId}/action`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ actionId })
  });
  return res.json();
}