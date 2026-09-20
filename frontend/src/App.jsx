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
import Learner from './pages/curator/Learner';
import ArenaIndex from './pages/arena/Index';
import ArenaPrepare from './pages/arena/Prepare';
import ArenaChat from './pages/arena/Chat';
import ArenaDebrief from './pages/arena/Debrief';
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
        <Route path="/curator/learner/:userId" element={<Learner />} />

        <Route path="/arena" element={<ArenaIndex />} />
        <Route path="/arena/prepare/:scenarioId" element={<ArenaPrepare />} />
        <Route path="/arena/chat/:scenarioId" element={<ArenaChat />} />
        <Route path="/arena/debrief/:scenarioId/:outcome" element={<ArenaDebrief />} />
      </Routes>
    </BrowserRouter>
  );
}