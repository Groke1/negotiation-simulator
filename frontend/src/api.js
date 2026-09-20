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
  localStorage.removeItem('user_id');
}

export function getToken() {
  return localStorage.getItem('access_token');
}

export function getUserId() {
  return localStorage.getItem('user_id');
}

// ==== ПОЛЬЗОВАТЕЛЬ ====

export async function getMe() {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/users/me`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки пользователя');
  return res.json();
}

export async function getRoles() {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/roles`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки ролей');
  return res.json();
}

export async function switchRole(role) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/roles/active`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ role })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка смены роли');
  }
  const data = await res.json();
  localStorage.setItem('access_token', data.access_token);
  return data;
}

export async function getOrganizations() {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/organizations`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки организаций');
  const data = await res.json();
  return data.organizations;
}

export async function setOrganization(organizationId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/users/me/organization`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ organization_id: organizationId })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка привязки организации');
  }
}

// ==== INTERNAL (для демо) ====

export async function grantRole(userId, role) {
  const res = await fetch(`http://localhost:8080/internal/v1/users/${userId}/roles`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ role })
  });
  if (!res.ok && res.status !== 409) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка выдачи роли');
  }
}

// ==== РЕГИСТРАЦИЯ С РОЛЬЮ ====

export async function registerWithRole(username, email, password, role) {
  const data = await register(username, email, password);

  if (role && role !== 'learner') {
    try {
      const me = await getMe();
      await grantRole(me.id, role);
    } catch (e) {
      console.warn('Не удалось выдать роль:', e.message);
    }
  }

  return data;
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

export async function getScenario(id) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/scenarios/${id}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Сценарий не найден');
  return res.json();
}

export async function generateScenario(prompt) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/scenarios/generate`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ prompt })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка генерации сценария');
  }
  return res.json();
}

export async function validateScenario(id) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/scenarios/${id}/validate`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка валидации');
  return res.json();
}

export async function publishScenario(id) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/scenarios/${id}/publish`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка публикации');
  }
  return res.json();
}

export async function getScenarioVersions(id) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/scenarios/${id}/versions`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки версий');
  const data = await res.json();
  return data.versions;
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

// ==== КУРАТОР ====

export async function getCuratorTeam() {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/curator/team`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (res.status === 403) throw new Error('Нужна роль куратора');
  if (!res.ok) throw new Error('Ошибка загрузки команды');
  const data = await res.json();
  return data.team;
}

export async function getCuratorAnalytics() {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/curator/team/analytics`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки аналитики');
  return res.json();
}

export async function getProfileHistory(userId, from, to) {
  const token = getToken();
  const url = new URL(`${BASE_URL}/users/${userId}/profile/history`);
  if (from) url.searchParams.set('from', from);
  if (to) url.searchParams.set('to', to);

  const res = await fetch(url.toString(), {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  if (!res.ok) throw new Error('Ошибка загрузки динамики');
  return res.json();
}

export async function assignTraining(learnerId, scenarioId) {
  const token = getToken();
  const res = await fetch(`${BASE_URL}/assignments`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ learner_id: learnerId, scenario_id: scenarioId })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка назначения');
  }
  return res.json();
}