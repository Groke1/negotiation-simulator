const KEY = 'arena_history';
const MAX_SESSIONS = 4; 

export function saveSession(session) {
  const all = getHistory();
  all.unshift({
    ...session,
    id: Date.now(),
    date: new Date().toLocaleString('ru-RU', {
      day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit'
    })
  });
  const trimmed = all.slice(0, MAX_SESSIONS);
  try {
    localStorage.setItem(KEY, JSON.stringify(trimmed));
  } catch (e) {
    console.error('Не удалось сохранить историю', e);
  }
}

export function getHistory() {
  try {
    return JSON.parse(localStorage.getItem(KEY) || '[]');
  } catch {
    return [];
  }
}

export function clearHistory() {
  localStorage.removeItem(KEY);
}