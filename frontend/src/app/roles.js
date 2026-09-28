export const roleHome = (role) => ({ learner: '/learner', methodist: '/methodist/scenarios', curator: '/curator/team' }[role] || '/learner');
export const roleName = (role) => ({ learner: 'Ученик', methodist: 'Методист', curator: 'Куратор' }[role] || role);
export const PRODUCT_ROLES = ['learner', 'methodist', 'curator'];
export const normalizeRoles = (roles = []) => roles.filter((role) => PRODUCT_ROLES.includes(role));
export const preferredRole = (roles = []) => {
  const availableRoles = normalizeRoles(roles);
  return ['learner', 'curator', 'methodist'].find((role) => availableRoles.includes(role));
};
export const resolveActiveRole = (me) => {
  const roles = normalizeRoles(me?.roles || []);
  if (PRODUCT_ROLES.includes(me?.active_role) && roles.includes(me.active_role)) return me.active_role;
  return preferredRole(roles) || 'learner';
};
