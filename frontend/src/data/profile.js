export const patternToCompetency = {
  ASK_QUESTION: 'diagnostics',
  MAKE_PROPOSAL: 'value',
  MAKE_CONCESSION: 'trade',
  REQUEST_RECIPROCITY: 'trade',
  HOLD_BOUNDARY: 'goals',
  PAUSE: 'conflict',
  SUMMARIZE: 'argumentation'
};

export const competencies = [
  { id: 'goals', name: 'Цели и границы' },
  { id: 'diagnostics', name: 'Диагностика' },
  { id: 'value', name: 'Создание ценности' },
  { id: 'argumentation', name: 'Аргументация' },
  { id: 'trade', name: 'Торг и обмен' },
  { id: 'conflict', name: 'Управление конфликтом' }
];

export function computeLevels(patterns) {
  const counts = {};
  competencies.forEach(c => counts[c.id] = 0);

  patterns.forEach(p => {
    const comp = patternToCompetency[p];
    if (comp) counts[comp]++;
  });

  const levels = {};
  competencies.forEach(c => {
    levels[c.id] = Math.min(4, counts[c.id]);
  });

  return levels;
}

export function computeOverallProfile(sessions) {
  if (sessions.length === 0) {
    const empty = {};
    competencies.forEach(c => empty[c.id] = 0);
    return empty;
  }

  const totals = {};
  competencies.forEach(c => totals[c.id] = 0);

  sessions.forEach(s => {
    const levels = computeLevels(s.patterns || []);
    competencies.forEach(c => {
      totals[c.id] += levels[c.id];
    });
  });

  const avg = {};
  competencies.forEach(c => {
    avg[c.id] = Math.min(4, Math.round(totals[c.id] / sessions.length));
  });

  return avg;
}

export const chartColors = [
  '#111111',
  '#2e6fff',
  '#0a7c2f',
  '#d97706'
];