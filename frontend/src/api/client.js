const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

const ERROR_MESSAGES = {
  ACCESS_TOKEN_INVALID: 'Сессия истекла. Войдите снова.',
  ACCESS_TOKEN_EXPIRED: 'Сессия истекла. Войдите снова.',
  ACCESS_TOKEN_MISSING: 'Войдите, чтобы продолжить.',
  FORBIDDEN: 'Недостаточно прав для этого действия.',
  ROLE_NOT_ASSIGNED: 'Эта роль не назначена вашему аккаунту.',
  SCENARIO_MODIFICATION_FORBIDDEN: 'Этот сценарий нельзя редактировать.',
  SCENARIO_IMMUTABLE: 'Опубликованный или архивный сценарий нельзя изменить.',
  SCENARIO_PUBLISH_VALIDATION_FAILED: 'Сценарий пока не готов к публикации.',
  ORGANIZATION_REQUIRED: 'Сначала выберите организацию.',
  ASSIGNMENT_CONFLICT: 'У ученика уже есть активное назначение этого сценария.',
  ARENA_JOIN_NOT_ALLOWED: 'Сейчас к этой Arena нельзя подключиться.',
  ARENA_MESSAGE_NOT_FOUND: 'Сообщение Arena не найдено.',
  MESSAGE_ALREADY_CLASSIFIED: 'Эта реплика уже размечена.',
  MESSAGE_NOT_CLASSIFIABLE: 'Эту реплику нельзя разметить.',
  ARENA_FEEDBACK_CONFLICT: 'Feedback этому участнику уже отправлен.',
  TRAINING_ANALYSIS_NOT_READY: 'Итоговый разбор тренировки ещё готовится.',
};

export class ApiError extends Error {
  constructor(message, { status = 0, code = 'NETWORK_ERROR', details } = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

let refreshInFlight = null;

export const tokenStore = {
  get: () => localStorage.getItem('access_token'),
  set: (token) => token && localStorage.setItem('access_token', token),
  clear: () => { localStorage.removeItem('access_token'); localStorage.removeItem('user_id'); },
};

async function parseResponse(response) {
  if (response.status === 204) return null;
  if (!(response.headers.get('content-type') || '').includes('application/json')) return null;
  return response.json().catch(() => null);
}

async function refreshToken() {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = fetch(`${API_BASE_URL}/auth/refresh`, { method: 'POST', credentials: 'include', headers: { Accept: 'application/json' } })
    .then(async (response) => {
      const body = await parseResponse(response);
      if (!response.ok || !body?.access_token) throw new ApiError('Сессия истекла. Войдите снова.', { status: response.status, code: body?.code || 'UNAUTHORIZED' });
      tokenStore.set(body.access_token);
      return body.access_token;
    }).finally(() => { refreshInFlight = null; });
  return refreshInFlight;
}

export async function request(path, { method = 'GET', data, body, signal, headers = {}, auth = true, retry = true } = {}) {
  const token = tokenStore.get();
  const payload = body ?? (data === undefined ? undefined : JSON.stringify(data));
  const requestHeaders = { Accept: 'application/json', ...(payload && !(payload instanceof FormData) ? { 'Content-Type': 'application/json' } : {}), ...(auth && token ? { Authorization: `Bearer ${token}` } : {}), ...headers };
  let response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, { method, body: payload, signal, credentials: 'include', headers: requestHeaders });
  } catch (error) {
    if (error.name === 'AbortError') throw error;
    throw new ApiError('Не удалось связаться с сервером. Проверьте подключение и повторите попытку.');
  }
  if (response.status === 401 && auth && retry) {
    try { await refreshToken(); return request(path, { method, data, body, signal, headers, auth, retry: false }); }
    catch (error) { tokenStore.clear(); window.dispatchEvent(new Event('app:unauthorized')); throw error; }
  }
  const result = await parseResponse(response);
  if (!response.ok) {
    const fallback = response.status === 403 ? 'Недостаточно прав для этого действия.' : response.status === 404 ? 'Данные не найдены.' : 'Не удалось выполнить действие. Попробуйте ещё раз.';
    const code = result?.code || `HTTP_${response.status}`;
    throw new ApiError(ERROR_MESSAGES[code] || fallback, { status: response.status, code, details: result });
  }
  return result;
}

export const apiBaseUrl = API_BASE_URL;

export function webSocketUrl(path) {
  const base = new URL(API_BASE_URL, window.location.origin);
  const protocol = base.protocol === 'https:' ? 'wss:' : 'ws:';
  return new URL(path, `${protocol}//${base.host}`).toString();
}
