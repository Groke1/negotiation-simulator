const BASE_URL = 'http://localhost:8080/api/v1';

// ==== АВТОРИЗАЦИЯ ====

export async function login(email, password) {
  const res = await fetch(`${BASE_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ email, password })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка входа');
  }
  const data = await res.json();
  localStorage.setItem('access_token', data.access_token);
  return data;
}

export async function register(username, email, password) {
  const res = await fetch(`${BASE_URL}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ username, email, password })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка регистрации');
  }
  const data = await res.json();
  localStorage.setItem('access_token', data.access_token);
  return data;
}

export function logout() {
  localStorage.removeItem('access_token');
}

export function getToken() {
  return localStorage.getItem('access_token');
}

// ==== СЦЕНАРИИ ====

export async function getScenarios() {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/scenarios`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (res.status === 401) throw new Error('Не авторизован');
  if (!res.ok) throw new Error('Ошибка загрузки сценариев');
  const data = await res.json();
  return data.scenarios;
}

// ==== ТРЕНИРОВКИ ====

export async function startSession(scenarioId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/sessions`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ scenario_id: scenarioId, channel: 'chat' })
  });
  if (!res.ok) throw new Error('Ошибка старта сессии');
  return res.json();
}

export async function getSessionState(sessionId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/sessions/${sessionId}/state`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки состояния');
  return res.json();
}

export async function sendAction(sessionId, actionId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/sessions/${sessionId}/events`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      action_id: actionId,
      client_event_id: crypto.randomUUID()
    })
  });
  if (!res.ok) throw new Error('Ошибка отправки действия');
  return res.json();
}

export async function completeSession(sessionId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/sessions/${sessionId}/complete`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка завершения');
  return res.json();
}

export async function getAnalysis(sessionId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/sessions/${sessionId}/analysis`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки разбора');
  return res.json();
}

// ==== ПРОФИЛЬ И РЕКОМЕНДАЦИЯ ====

export async function getProfile(userId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/users/${userId}/profile`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки профиля');
  return res.json();
}

export async function getRecommendation(userId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/users/${userId}/recommendation`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (res.status === 204) return null;
  if (!res.ok) throw new Error('Ошибка загрузки рекомендации');
  return res.json();
}

export async function getSessionHistory(userId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/users/${userId}/sessions`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки истории');
  return res.json();
}