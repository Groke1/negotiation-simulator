import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import Chat from './pages/Chat';
import Debrief from './pages/Debrief';
import Profile from './pages/Profile';
import History from './pages/History';
import Login from './pages/Login';
import Constructor from './pages/constructor/Index';
import CreateScenario from './pages/constructor/Create';
import ScenarioView from './pages/constructor/Scenario';
import Curator from './pages/curator/Index';
import CuratorHistory from './pages/curator/History';
import Learner from './pages/curator/Learner';
import CuratorArena from './pages/curator/Arena';
import ArenaFeedback from './pages/curator/ArenaFeedback';
import ArenaIndex from './pages/arena/Index';
import ArenaSession from './pages/arena/Session';
import ArenaHistory from './pages/arena/History';
import './App.css';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route path="/home" element={<Home />} />
        <Route path="/chat/:scenarioId" element={<Chat />} />
        <Route path="/debrief/:scenarioId/:outcome" element={<Debrief />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/history" element={<History />} />

        <Route path="/constructor" element={<Constructor />} />
        <Route path="/constructor/create" element={<CreateScenario />} />
        <Route path="/constructor/scenario/:scenarioId" element={<ScenarioView />} />

        <Route path="/curator" element={<Curator />} />
        <Route path="/curator/history" element={<CuratorHistory />} />
        <Route path="/curator/learner/:userId" element={<Learner />} />
        <Route path="/curator/arena" element={<CuratorArena />} />
        <Route path="/curator/arena/:sessionId/feedback" element={<ArenaFeedback />} />

        <Route path="/arena" element={<ArenaIndex />} />
        <Route path="/arena/assignment/:assignmentId/join" element={<ArenaSession />} />
        <Route path="/arena/session/:sessionId" element={<ArenaSession />} />
        <Route path="/arena/history/:sessionId" element={<ArenaHistory />} />
      </Routes>
    </BrowserRouter>
  );
}