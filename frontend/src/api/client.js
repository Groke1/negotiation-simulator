const BASE_URL = 'http://localhost:8080/api/v1';
export const tokenStore = { get: () => localStorage.getItem('access_token'), set: (token) => localStorage.setItem('access_token', token), clear: () => { localStorage.removeItem('access_token'); localStorage.removeItem('user_id'); } };
export const webSocketUrl = (path) => `ws://localhost:8080${path}`;
export async function request(path, { method = 'GET', data, auth = true, signal } = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (auth && tokenStore.get()) headers.Authorization = `Bearer ${tokenStore.get()}`;
  const response = await fetch(`${BASE_URL}${path}`, { method, headers, body: data ? JSON.stringify(data) : undefined, signal, credentials: 'include' });
  const text = await response.text();
  const body = text ? JSON.parse(text) : null;
  if (response.status === 401) { tokenStore.clear(); window.dispatchEvent(new CustomEvent('app:unauthorized')); }
  if (!response.ok) { const error = new Error(body?.message || 'Ошибка запроса'); error.status = response.status; error.details = body; throw error; }
  return body;
}