import { request, tokenStore } from './api/client';

const query = (values = {}) => {
  const params = new URLSearchParams(Object.entries(values).filter(([, value]) => value !== undefined && value !== null && value !== ''));
  return params.size ? `?${params}` : '';
};

export const login = (email, password, signal) => request('/auth/login', { method: 'POST', data: { email, password }, auth: false, signal }).then((data) => { tokenStore.set(data.access_token); return data; });
export const register = (username, email, password, signal) => request('/auth/register', { method: 'POST', data: { username, email, password }, auth: false, signal }).then((data) => { tokenStore.set(data.access_token); return data; });
export const logout = async () => { try { await request('/auth/logout', { method: 'POST' }); } finally { tokenStore.clear(); } };
export const getToken = tokenStore.get;
export const getUserId = () => localStorage.getItem('user_id');
export const getMe = (signal) => request('/users/me', { signal });
export const getRoles = (signal) => request('/roles', { signal });
export const grantRole = (role) => request('/roles', { method: 'POST', data: { role } });
export const switchRole = (role) => request('/roles/active', { method: 'PUT', data: { role } }).then((data) => { if (data?.access_token) tokenStore.set(data.access_token); return data; });
export const getOrganizations = (signal) => request('/organizations', { signal }).then((data) => data.organizations || []);
export const createOrganization = (name) => request('/organizations', { method: 'POST', data: { name } });
export const setOrganization = (organizationId) => request('/users/me/organization', { method: 'PUT', data: { organization_id: organizationId } });

export const getScenarios = (filters, signal) => request(`/scenarios${query(filters)}`, { signal }).then((data) => data.scenarios || []);
export const getScenario = (id, signal) => request(`/scenarios/${id}`, { signal });
export const createScenario = (data) => request('/scenarios', { method: 'POST', data });
export const updateScenario = (id, data) => request(`/scenarios/${id}`, { method: 'PUT', data });
export const validateScenario = (id) => request(`/scenarios/${id}/validate`, { method: 'POST' });
export const publishScenario = (id) => request(`/scenarios/${id}/publish`, { method: 'POST' });
export const archiveScenario = (id) => request(`/scenarios/${id}/archive`, { method: 'POST' });
export const cloneScenario = (id) => request(`/scenarios/${id}/clone`, { method: 'POST' });
export const setScenarioFavorite = (id, favorite) => request(`/scenarios/${id}/favorite`, { method: favorite ? 'PUT' : 'DELETE' });
export const getSemanticActions = (signal) => request('/semantic-actions', { signal }).then((data) => data.actions || []);
export const getCompetencyClusters = (signal) => request('/competency-clusters', { signal }).then((data) => data.clusters || []);

export const generateScenarioContext = (scenarioId, instructions) => request(`/scenarios/${scenarioId}/ai/expand-context`, { method: 'POST', data: { instructions } });
export const generateScenarioDraft = (scenarioId, instructions) => request(`/scenarios/${scenarioId}/ai/generate-draft`, { method: 'POST', data: { instructions } });
export const generateScenarioBlock = (scenarioId, blockId, instructions) => request(`/scenarios/${scenarioId}/blocks/${blockId}/ai/generate`, { method: 'POST', data: { instructions } });
export const generateDialogue = (scenarioId, blockId, learner_role_id, instructions) => request(`/scenarios/${scenarioId}/blocks/${blockId}/ai/generate-dialogue`, { method: 'POST', data: { learner_role_id, instructions } });
export const getAIGenerationRequest = (scenarioId, requestId, signal) => request(`/scenarios/${scenarioId}/ai/requests/${requestId}`, { signal });

export const startTrainingSession = (payload) => request('/training/sessions', { method: 'POST', data: payload });
export const getTrainingSessions = (signal) => request('/training/sessions', { signal }).then((data) => data.sessions || []);
export const getTrainingSession = (id, signal) => request(`/training/sessions/${id}`, { signal });
export const sendTrainingMessage = (id, message) => request(`/training/sessions/${id}/messages`, { method: 'POST', data: message });
export const sendTrainingAction = (id, action) => request(`/training/sessions/${id}/actions`, { method: 'POST', data: action });
export const classifyPartnerMessage = (sessionId, messageId, semantic_action_code) => request(`/training/sessions/${sessionId}/messages/${messageId}/classification`, { method: 'POST', data: { semantic_action_code } });
export const getTrainingAnalysis = (id, signal) => request(`/training/sessions/${id}/analysis`, { signal });

export const getMyProfile = (signal) => request('/users/me/profile', { signal });
export const getUserProfile = (id, signal) => request(`/users/${id}/profile`, { signal });
export const getProfileHistory = (id, signal) => request(`/users/${id}/profile/history`, { signal });
export const getOrganizationUsers = (filters, signal) => request(`/organizations/me/users${query(filters)}`, { signal }).then((data) => data.users || []);
export const getCuratorTeam = (signal) => request('/curator/team', { signal }).then((data) => data.team || []);
export const getCuratorTeamSessions = (filters, signal) => request(`/curator/team/sessions${query(filters)}`, { signal }).then((data) => data.sessions || []);
export const getCuratorAnalytics = (signal) => request('/curator/team/analytics', { signal });
export const createAssignment = (learnerId, scenarioId) => request('/assignments', { method: 'POST', data: { learner_id: learnerId, scenario_id: scenarioId } });
export const getAssignments = (filters, signal) => request(`/assignments${query(filters)}`, { signal }).then((data) => data.assignments || []);

export const getArenaAssignments = (filters, signal) => request(`/arena/assignments${query(filters)}`, { signal }).then((data) => data.arenas || []);
export const createArenaAssignment = (data, participants, scheduledAt) => request('/arena/assignments', { method: 'POST', data: typeof data === 'string' ? { scenario_id: data, participants, scheduled_at: scheduledAt } : data });
export const getArenaAssignment = (id, signal) => request(`/arena/assignments/${id}`, { signal });
export const joinArena = (id) => request(`/arena/assignments/${id}/join`, { method: 'POST' });
export const getArenaSession = (id, signal) => request(`/arena/sessions/${id}`, { signal });
export const getArenaMessages = (id, signal) => request(`/arena/sessions/${id}/messages`, { signal }).then((data) => data.messages || []);
export const sendArenaMessage = (id, message) => request(`/arena/sessions/${id}/messages`, { method: 'POST', data: message });
export const classifyArenaMessage = (sessionId, messageId, data) => request(`/arena/sessions/${sessionId}/messages/${messageId}/classification`, { method: 'POST', data });
export const completeArena = (id) => request(`/arena/sessions/${id}/complete`, { method: 'POST' });
export const getArenaFeedback = (id, signal) => request(`/arena/sessions/${id}/feedback`, { signal }).then((data) => data.feedback || []);
export const createArenaFeedback = (sessionId, participantId, comment, ratings = []) => request(`/arena/sessions/${sessionId}/feedback`, { method: 'POST', data: { participant_id: participantId, comment, ratings } });
export const getMyArenaAssignments = (id, signal) => request(`/users/${id}/arena-assignments`, { signal }).then((data) => data.arenas || []);
export const getMyArenaHistory = (id, signal) => request(`/users/${id}/arena-history`, { signal }).then((data) => data.arenas || []);
export const getArenaHistoryDetail = (userId, sessionId, signal) => request(`/users/${userId}/arena-history/${sessionId}`, { signal });
