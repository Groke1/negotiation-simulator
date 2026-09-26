import { BrowserRouter, Navigate, Outlet, Route, Routes, useLocation } from 'react-router-dom';
import { AuthProvider, useAuth } from './app/AuthContext';
import { resolveActiveRole, roleHome } from './app/roles';
import AppShell from './components/AppShell';
import { LoadingState } from './components/UI';
import Login from './pages/Login';
import Home from './pages/Home';
import Chat from './pages/Chat';
import Debrief from './pages/Debrief';
import Profile from './pages/Profile';
import History from './pages/History';
import Constructor from './pages/constructor/Index';
import CreateScenario from './pages/constructor/Create';
import ScenarioView from './pages/constructor/Scenario';
import Curator from './pages/curator/Index';
import Learner from './pages/curator/Learner';
import CuratorArena from './pages/curator/Arena';
import ArenaFeedback from './pages/curator/ArenaFeedback';
import CuratorHistory from './pages/curator/History';
import ArenaIndex from './pages/arena/Index';
import ArenaHistory from './pages/arena/History';
import ArenaSession from './pages/arena/Session';
import Onboarding from './pages/Onboarding';
import './App.css';

function Protected() { const { authenticated, loading } = useAuth(); const location = useLocation(); if (loading) return <LoadingState />; return authenticated ? <Outlet /> : <Navigate to="/auth/login" replace state={{ from: location }} />; }
function WorkspaceReady() { const { me } = useAuth(); if (!me?.organization || !me?.roles?.length) return <Navigate to="/onboarding" replace />; return <Outlet />; }
function Role({ role }) { const { me } = useAuth(); const activeRole = resolveActiveRole(me); return activeRole === role ? <Outlet /> : <Navigate to={roleHome(activeRole)} replace />; }
function RootRedirect() { const { authenticated, loading, me } = useAuth(); if (loading) return <LoadingState />; return <Navigate to={authenticated ? (!me?.organization || !me?.roles?.length ? '/onboarding' : roleHome(resolveActiveRole(me))) : '/auth/login'} replace />; }

export default function App() { return <BrowserRouter><AuthProvider><Routes>
  <Route path="/auth/login" element={<Login />} /><Route path="/auth/register" element={<Login register />} /><Route path="/" element={<RootRedirect />} />
  <Route element={<Protected />}><Route path="/onboarding" element={<Onboarding />} /><Route element={<WorkspaceReady />}><Route element={<AppShell />}>
    <Route element={<Role role="learner" />}><Route path="/learner" element={<Home />} /><Route path="/learner/training" element={<Home catalog />} /><Route path="/learner/training/:sessionId" element={<Chat />} /><Route path="/learner/training/:sessionId/analysis" element={<Debrief />} /><Route path="/learner/profile" element={<Profile />} /><Route path="/learner/history" element={<History />} /><Route path="/learner/arenas" element={<ArenaIndex />} /><Route path="/learner/arenas/join/:assignmentId" element={<ArenaSession />} /><Route path="/learner/arenas/:sessionId" element={<ArenaHistory />} /></Route>
    <Route element={<Role role="methodist" />}><Route path="/methodist/scenarios" element={<Constructor />} /><Route path="/methodist/scenarios/new" element={<CreateScenario />} /><Route path="/methodist/scenarios/:scenarioId" element={<ScenarioView />} /></Route>
    <Route element={<Role role="curator" />}><Route path="/curator/team" element={<Curator />} /><Route path="/curator/team/:userId" element={<Learner />} /><Route path="/curator/training/:sessionId/analysis" element={<Debrief />} /><Route path="/curator/assignments" element={<CuratorHistory assignments />} /><Route path="/curator/arenas" element={<CuratorArena />} /><Route path="/curator/arenas/:sessionId" element={<ArenaFeedback />} /><Route path="/curator/history" element={<CuratorHistory />} /><Route path="/curator/analytics" element={<Curator />} /></Route>
  </Route></Route></Route>
  <Route path="*" element={<Navigate to="/" replace />} />
</Routes></AuthProvider></BrowserRouter>; }
