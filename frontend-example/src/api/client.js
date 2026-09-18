const baseURL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";
let accessToken = sessionStorage.getItem("arena_access_token") || "";

function setToken(token) {
  accessToken = token;
  if (token) sessionStorage.setItem("arena_access_token", token);
  else sessionStorage.removeItem("arena_access_token");
}

function queryString(params = {}) {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") search.set(key, value);
  });
  const value = search.toString();
  return value ? `?${value}` : "";
}

async function request(path, options = {}, retry = true) {
  const headers = new Headers(options.headers);
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
  if (options.body) headers.set("Content-Type", "application/json");

  const response = await fetch(`${baseURL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  if (response.status === 401 && retry && accessToken) {
    const refreshed = await fetch(`${baseURL}/auth/refresh`, {
      method: "POST",
      credentials: "include",
    });
    if (refreshed.ok) {
      const data = await refreshed.json();
      setToken(data.access_token);
      return request(path, options, false);
    }
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    const message = error.message || error.detail || `Ошибка API (${response.status})`;
    const requestError = new Error(message);
    requestError.status = response.status;
    requestError.code = error.code;
    throw requestError;
  }

  if (response.status === 204) return null;
  return response.json();
}

export const api = {
  baseURL,
  token: () => accessToken,
  clearToken: () => setToken(""),

  async register(payload) {
    const data = await request("/auth/register", { method: "POST", body: JSON.stringify(payload) });
    setToken(data.access_token);
    return data;
  },
  async login(payload) {
    const data = await request("/auth/login", { method: "POST", body: JSON.stringify(payload) });
    setToken(data.access_token);
    return data;
  },
  async logout() {
    try {
      await request("/auth/logout", { method: "POST" });
    } finally {
      setToken("");
    }
  },

  me: () => request("/users/me"),
  organizations: () => request("/organizations"),
  createOrganization: (name) => request("/organizations", { method: "POST", body: JSON.stringify({ name }) }),
  setOrganization: (organization_id) => request("/users/me/organization", { method: "PUT", body: JSON.stringify({ organization_id }) }),
  grantRole: (role) => request("/roles", { method: "POST", body: JSON.stringify({ role }) }),
  switchRole: async (role) => {
    const data = await request("/roles/active", { method: "PUT", body: JSON.stringify({ role }) });
    setToken(data.access_token);
    return data;
  },

  scenarios: () => request("/scenarios"),
  scenario: (id) => request(`/scenarios/${id}`),
  createScenario: (payload) => request("/scenarios", { method: "POST", body: JSON.stringify(payload) }),
  updateScenario: (id, payload) => request(`/scenarios/${id}`, { method: "PUT", body: JSON.stringify(payload) }),
  generateScenario: (prompt) => request("/scenarios/generate", { method: "POST", body: JSON.stringify({ prompt }) }),
  validateScenario: (id) => request(`/scenarios/${id}/validate`, { method: "POST" }),
  publishScenario: (id) => request(`/scenarios/${id}/publish`, { method: "POST" }),
  archiveScenario: (id) => request(`/scenarios/${id}/archive`, { method: "POST" }),
  scenarioVersions: (id) => request(`/scenarios/${id}/versions`),
  scenarioVersion: (id, version) => request(`/scenarios/${id}/versions/${encodeURIComponent(version)}`),

  organizationUsers: () => request("/organizations/me/users"),
  curatorTeam: (params) => request(`/curator/team${queryString(params)}`),
  teamAnalytics: () => request("/curator/team/analytics"),
  profileHistory: (userID, params) => request(`/users/${userID}/profile/history${queryString(params)}`),
  createTrainingAssignment: (payload) => request("/assignments", { method: "POST", body: JSON.stringify(payload) }),

  createArena: (payload) => request("/arena/assignments", { method: "POST", body: JSON.stringify(payload) }),
  organizationArenas: () => request("/arena/assignments"),
  arenaAssignment: (id) => request(`/arena/assignments/${id}`),
  arenaAssignments: (userID) => request(`/users/${userID}/arena-assignments`),
  arenaHistory: (userID) => request(`/users/${userID}/arena-history`),
  arenaHistoryDetails: (userID, sessionID) => request(`/users/${userID}/arena-history/${sessionID}`),
  joinArena: (id) => request(`/arena/assignments/${id}/join`, { method: "POST" }),
  arenaSession: (id) => request(`/arena/sessions/${id}`),
  arenaMessages: (id) => request(`/arena/sessions/${id}/messages`),
  completeArena: (id) => request(`/arena/sessions/${id}/complete`, { method: "POST" }),
  arenaFeedback: (id) => request(`/arena/sessions/${id}/feedback`),
  createFeedback: (id, payload) => request(`/arena/sessions/${id}/feedback`, { method: "POST", body: JSON.stringify(payload) }),

  startTraining: (payload) => request("/training/sessions", { method: "POST", body: JSON.stringify(payload) }),
  trainingSession: (id) => request(`/training/sessions/${id}`),
  trainingAction: (id, payload) => request(`/training/sessions/${id}/actions`, { method: "POST", body: JSON.stringify(payload) }),
  trainingMessage: (id, payload) => request(`/training/sessions/${id}/messages`, { method: "POST", body: JSON.stringify(payload) }),
  completeTraining: (id) => request(`/training/sessions/${id}/complete`, { method: "POST" }),
  trainingAnalysis: (id) => request(`/training/sessions/${id}/analysis`),

  arenaWebSocketURL: (sessionID, websocketPath) => {
    const url = new URL(baseURL);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    url.pathname = websocketPath || `/ws/arena/sessions/${sessionID}`;
    url.search = "";
    return url.toString();
  },
};
