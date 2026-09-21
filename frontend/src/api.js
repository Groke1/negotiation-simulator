const BASE_URL = 'http://localhost:8080/api/v1';

// ========================================
// ==== ОБЁРТКА fetch с автоматическим refresh ====
// ========================================

let isRefreshing = false;
let refreshPromise = null;

async function refreshAccessToken() {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const res = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include'
      });

      if (!res.ok) {
        // Refresh не сработал — разлогиниваем
        localStorage.removeItem('access_token');
        localStorage.removeItem('user_id');
        window.location.href = '/';
        throw new Error('Сессия истекла, войдите заново');
      }

      const data = await res.json();
      localStorage.setItem('access_token', data.access_token);
      return data.access_token;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

async function fetchWithAuth(url, options = {}) {
  const token = localStorage.getItem('access_token');

  const headers = {
    ...(options.headers || {}),
    'Authorization': `Bearer ${token}`
  };

  let res = await fetch(url, { ...options, headers });

  // Если 401 — пробуем refresh и повторяем
  if (res.status === 401) {
    try {
      const newToken = await refreshAccessToken();
      const retryHeaders = {
        ...(options.headers || {}),
        'Authorization': `Bearer ${newToken}`
      };
      res = await fetch(url, { ...options, headers: retryHeaders });
    } catch (err) {
      throw new Error('Не удалось обновить сессию');
    }
  }

  return res;
}

// ========================================
// ==== АВТОРИЗАЦИЯ ====
// ========================================

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

// ========================================
// ==== ПОЛЬЗОВАТЕЛЬ ====
// ========================================

export async function getMe() {
  const res = await fetchWithAuth(`${BASE_URL}/users/me`);
  if (!res.ok) throw new Error('Ошибка загрузки пользователя');
  return res.json();
}

export async function getRoles() {
  const res = await fetchWithAuth(`${BASE_URL}/roles`);
  if (!res.ok) throw new Error('Ошибка загрузки ролей');
  return res.json();
}

export async function switchRole(role) {
  const res = await fetchWithAuth(`${BASE_URL}/roles/active`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
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
  const res = await fetchWithAuth(`${BASE_URL}/organizations`);
  if (!res.ok) throw new Error('Ошибка загрузки организаций');
  const data = await res.json();
  return data.organizations;
}

export async function setOrganization(organizationId) {
  const res = await fetchWithAuth(`${BASE_URL}/users/me/organization`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ organization_id: organizationId })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка привязки организации');
  }
}

// ========================================
// ==== INTERNAL (для демо) ====
// ========================================

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

// ========================================
// ==== СЦЕНАРИИ ====
// ========================================

export async function getScenarios() {
  const res = await fetchWithAuth(`${BASE_URL}/scenarios`);
  if (res.status === 401) throw new Error('Не авторизован');
  if (!res.ok) throw new Error('Ошибка загрузки сценариев');
  const data = await res.json();
  return data.scenarios || [];
}

export async function getScenario(id) {
  const res = await fetchWithAuth(`${BASE_URL}/scenarios/${id}`);
  if (!res.ok) throw new Error('Сценарий не найден');
  return res.json();
}

export async function updateScenario(scenarioId, scenarioData) {
  const res = await fetchWithAuth(`${BASE_URL}/scenarios/${scenarioId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(scenarioData)
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка сохранения');
  }
  return res.json();
}

export async function generateScenario(prompt) {
  const res = await fetchWithAuth(`${BASE_URL}/scenarios/generate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prompt })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка генерации сценария');
  }
  return res.json();
}

export async function validateScenario(id) {
  const res = await fetchWithAuth(`${BASE_URL}/scenarios/${id}/validate`, {
    method: 'POST'
  });
  if (!res.ok) throw new Error('Ошибка валидации');
  return res.json();
}

export async function publishScenario(id) {
  const res = await fetchWithAuth(`${BASE_URL}/scenarios/${id}/publish`, {
    method: 'POST'
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка публикации');
  }
  return res.json();
}

export async function getScenarioVersions(id) {
  const res = await fetchWithAuth(`${BASE_URL}/scenarios/${id}/versions`);
  if (!res.ok) throw new Error('Ошибка загрузки версий');
  const data = await res.json();
  return data.versions;
}

// ========================================
// ==== ТРЕНИРОВКИ ====
// ========================================

export async function startSession(scenarioId) {
  const res = await fetchWithAuth(`${BASE_URL}/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ scenario_id: scenarioId, channel: 'chat' })
  });
  if (!res.ok) throw new Error('Ошибка старта сессии');
  return res.json();
}

export async function getSessionState(sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/sessions/${sessionId}/state`);
  if (!res.ok) throw new Error('Ошибка загрузки состояния');
  return res.json();
}

export async function sendAction(sessionId, actionId) {
  const res = await fetchWithAuth(`${BASE_URL}/sessions/${sessionId}/events`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      action_id: actionId,
      client_event_id: crypto.randomUUID()
    })
  });
  if (!res.ok) throw new Error('Ошибка отправки действия');
  return res.json();
}

export async function completeSession(sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/sessions/${sessionId}/complete`, {
    method: 'POST'
  });
  if (!res.ok) throw new Error('Ошибка завершения');
  return res.json();
}

export async function getAnalysis(sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/sessions/${sessionId}/analysis`);
  if (!res.ok) throw new Error('Ошибка загрузки разбора');
  return res.json();
}

// ========================================
// ==== ПРОФИЛЬ И РЕКОМЕНДАЦИЯ ====
// ========================================

export async function getProfile(userId) {
  const res = await fetchWithAuth(`${BASE_URL}/users/${userId}/profile`);
  if (!res.ok) throw new Error('Ошибка загрузки профиля');
  return res.json();
}

export async function getRecommendation(userId) {
  const res = await fetchWithAuth(`${BASE_URL}/users/${userId}/recommendation`);
  if (res.status === 204) return null;
  if (!res.ok) throw new Error('Ошибка загрузки рекомендации');
  return res.json();
}

export async function getSessionHistory(userId) {
  const res = await fetchWithAuth(`${BASE_URL}/users/${userId}/sessions`);
  if (!res.ok) throw new Error('Ошибка загрузки истории');
  return res.json();
}

// ========================================
// ==== КУРАТОР ====
// ========================================

export async function getCuratorTeam() {
  const res = await fetchWithAuth(`${BASE_URL}/curator/team`);
  if (res.status === 403) throw new Error('Нужна роль куратора');
  if (!res.ok) throw new Error('Ошибка загрузки команды');
  const data = await res.json();
  return data.team;
}

export async function getCuratorAnalytics() {
  const res = await fetchWithAuth(`${BASE_URL}/curator/team/analytics`);
  if (!res.ok) throw new Error('Ошибка загрузки аналитики');
  return res.json();
}

export async function getProfileHistory(userId, from, to) {
  const url = new URL(`${BASE_URL}/users/${userId}/profile/history`);
  if (from) url.searchParams.set('from', from);
  if (to) url.searchParams.set('to', to);

  const res = await fetchWithAuth(url.toString());
  if (!res.ok) throw new Error('Ошибка загрузки динамики');
  return res.json();
}

export async function assignTraining(learnerId, scenarioId) {
  const res = await fetchWithAuth(`${BASE_URL}/assignments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ learner_id: learnerId, scenario_id: scenarioId })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка назначения');
  }
  return res.json();
}

// ========================================
// ==== АРЕНА ====
// ========================================

export async function getMyArenaAssignments(userId) {
  const res = await fetchWithAuth(`${BASE_URL}/users/${userId}/arena-assignments`);
  if (!res.ok) throw new Error('Ошибка загрузки арен');
  const data = await res.json();
  return data.arenas;
}

export async function getMyArenaHistory(userId) {
  const res = await fetchWithAuth(`${BASE_URL}/users/${userId}/arena-history`);
  if (!res.ok) throw new Error('Ошибка загрузки истории арен');
  const data = await res.json();
  return data.arenas;
}

export async function getArenaAssignment(assignmentId) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/assignments/${assignmentId}`);
  if (!res.ok) throw new Error('Арена не найдена');
  return res.json();
}

export async function joinArena(assignmentId) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/assignments/${assignmentId}/join`, {
    method: 'POST'
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Не удалось подключиться');
  }
  return res.json();
}

export async function getArenaSession(sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/sessions/${sessionId}`);
  if (!res.ok) throw new Error('Сессия не найдена');
  return res.json();
}

export async function getArenaMessages(sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/sessions/${sessionId}/messages`);
  if (!res.ok) throw new Error('Ошибка загрузки сообщений');
  const data = await res.json();
  return data.messages;
}

export async function completeArena(sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/sessions/${sessionId}/complete`, {
    method: 'POST'
  });
  if (!res.ok) throw new Error('Ошибка завершения');
}

export async function getArenaFeedback(sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/sessions/${sessionId}/feedback`);
  if (!res.ok) throw new Error('Ошибка загрузки фидбека');
  const data = await res.json();
  return data.feedback;
}

export async function createArenaAssignment(scenarioId, participants, scheduledAt) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/assignments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      scenario_id: scenarioId,
      participants,
      scheduled_at: scheduledAt
    })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка назначения арены');
  }
  return res.json();
}

export async function getCuratorArenas() {
  const res = await fetchWithAuth(`${BASE_URL}/arena/assignments`);
  if (res.status === 403) throw new Error('Нужна роль куратора');
  if (!res.ok) throw new Error('Ошибка загрузки арен');
  const data = await res.json();
  return data.arenas;
}

export async function createArenaFeedback(sessionId, participantId, comment, ratings = []) {
  const res = await fetchWithAuth(`${BASE_URL}/arena/sessions/${sessionId}/feedback`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ participant_id: participantId, comment, ratings })
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.message || 'Ошибка сохранения фидбека');
  }
  return res.json();
}

export async function getArenaHistoryDetail(userId, sessionId) {
  const res = await fetchWithAuth(`${BASE_URL}/users/${userId}/arena-history/${sessionId}`);
  if (!res.ok) throw new Error('Арена не найдена');
  return res.json();
}