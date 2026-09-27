import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { useAuth } from '../app/AuthContext';
import { resolveActiveRole, roleHome, roleName } from '../app/roles';
import ErrorBoundary from './ErrorBoundary';
import './user-popover.css';

const NAV = {
  learner: [['/learner', 'Главная'], ['/learner/training', 'Тренировки'], ['/learner/arenas', 'Арены'], ['/learner/history', 'История'], ['/learner/profile', 'Профиль']],
  methodist: [['/methodist/scenarios', 'Сценарии']],
  curator: [['/curator/team', 'Команда'], ['/curator/assignments', 'Назначения'], ['/curator/arenas', 'Арены'], ['/curator/history', 'История']],
};
export default function AppShell() {
  const { me, setRole, signOut } = useAuth(); const navigate = useNavigate(); const location = useLocation(); const [open, setOpen] = useState(false); const [userOpen, setUserOpen] = useState(false); const [switching, setSwitching] = useState(false);
  const activeRole = resolveActiveRole(me);
  async function changeRole(role) { if (role === activeRole) return setOpen(false); setSwitching(true); try { await setRole(role); navigate(roleHome(role)); } finally { setSwitching(false); setOpen(false); } }
  return <div className="app-shell"><header className="app-header"><NavLink className="brand" to={roleHome(activeRole)}>Арена <b>переговоров</b></NavLink><span className="organization">{me?.organization?.name || 'Организация не выбрана'}</span><div className="role-menu"><button className="role-trigger" onClick={() => setOpen(!open)} disabled={switching}>{roleName(activeRole)} ▾</button>{open && <div className="role-popover">{(me?.roles || []).map((role) => <button key={role} onClick={() => changeRole(role)}>{roleName(role)} {role === activeRole && '✓'}</button>)}</div>}</div><div className="user-profile"><button className="user-menu" onClick={() => setUserOpen((value) => !value)}>{me?.username || 'Профиль'} <span>▾</span></button>{userOpen && <div className="user-popover"><strong>{me?.username || 'Пользователь'}</strong><span>{me?.email}</span><div><small>Организация</small><b>{me?.organization?.name || 'Не выбрана'}</b></div><div><small>Роли</small><b>{(me?.roles || []).map(roleName).join(' · ') || 'Не назначены'}</b></div><hr /><button className="logout-button" onClick={() => { signOut(); navigate('/auth/login'); }}>Выйти</button></div>}</div></header><div className="shell-body"><nav className="side-nav">{(NAV[activeRole] || []).map(([to, label]) => <NavLink key={to} to={to} end={to === roleHome(activeRole)}>{label}</NavLink>)}</nav><main className="main-content"><ErrorBoundary resetKey={`${activeRole}:${location.pathname}`}><Outlet /></ErrorBoundary></main></div></div>;
}
