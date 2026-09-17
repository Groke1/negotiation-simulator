import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import Chat from './pages/Chat';
import Debrief from './pages/Debrief';
import Profile from './pages/Profile';
import History from './pages/History';
import Login from './pages/Login';
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
      </Routes>
    </BrowserRouter>
  );
}