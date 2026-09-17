import { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario } from '../data/scenario';
import { patterns, detectPattern, getPatternLabel } from '../data/patterns';
import { saveSession } from '../data/history';

export default function Chat() {
  const { scenarioId } = useParams();
  const navigate = useNavigate();
  const scenario = getScenario(scenarioId);

  const [currentNodeId, setCurrentNodeId] = useState('n1');
  const [history, setHistory] = useState([]);
  const [isTyping, setIsTyping] = useState(false);
  const [selectedPattern, setSelectedPattern] = useState(null);
  const [text, setText] = useState('');
  const [usedPatterns, setUsedPatterns] = useState([]);
  const scrollRef = useRef(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [history, isTyping]);

  if (!scenario) {
    return (
      <div className="page">
        <h1>Сценарий не найден</h1>
        <button onClick={() => navigate('/')}>На главную</button>
      </div>
    );
  }

  const node = scenario.nodes.find(n => n.id === currentNodeId);

  function chooseAction(action) {
    if (!action) return;

    const userText = text.trim() || action.label;
    const pattern = selectedPattern || action.pattern;

    const newHistory = [
      ...history,
      { partner: node.text, user: userText, pattern }
    ];
    setHistory(newHistory);
    setUsedPatterns([...usedPatterns, pattern]);
    setText('');
    setSelectedPattern(null);

    if (action.next.startsWith('end_')) {
      saveSession({
        scenarioId: scenario.id,
        scenarioTitle: scenario.title,
        outcome: action.next,
        role: scenario.roles.learner,
        patterns: [...usedPatterns, pattern]
      });
      navigate(`/debrief/${scenario.id}/${action.next}`);
      return;
    }

    setIsTyping(true);
    setTimeout(() => {
      setCurrentNodeId(action.next);
      setIsTyping(false);
    }, 1500);
  }

  function handleSend() {
    if (!text.trim() && !selectedPattern) return;

    let action;

    if (selectedPattern) {
      action = node.actions.find(a => a.pattern === selectedPattern);
      if (!action) action = node.actions[0];
    } else {
      const detected = detectPattern(text);
      action = node.actions.find(a => a.pattern === detected) || node.actions[0];
    }

    chooseAction(action);
  }

  return (
    <div className="chat-page">
      <header className="chat-header">
        <button className="back-btn" onClick={() => navigate('/')} aria-label="Назад">←</button>
        <div className="chat-header-info">
          <h2>{scenario.title}</h2>
          <p>{scenario.roles.learner} · {scenario.roles.partner} · {scenario.difficulty}</p>
        </div>
      </header>

      <div className="chat-stream" ref={scrollRef}>
        {history.map((h, i) => (
          <div key={i} className="msg-group">
            <div className="msg msg-partner">
              <span className="msg-sender">{scenario.roles.partner}</span>
              <div className="msg-bubble">{h.partner}</div>
            </div>
            <div className="msg msg-user">
              <span className="msg-sender">Вы</span>
              <div className="msg-bubble">{h.user}</div>
              {h.pattern && <span className="msg-pattern">{getPatternLabel(h.pattern)}</span>}
            </div>
          </div>
        ))}

        {node && !isTyping && (
          <div className="msg msg-partner">
            <span className="msg-sender">{scenario.roles.partner}</span>
            <div className="msg-bubble">{node.text}</div>
          </div>
        )}

        {isTyping && (
          <div className="msg msg-partner">
            <span className="msg-sender">{scenario.roles.partner}</span>
            <div className="msg-bubble typing">
              <span></span><span></span><span></span>
              <em>печатает</em>
            </div>
          </div>
        )}
      </div>

      <div className="chat-input-panel">
        <div className="patterns-row">
          {patterns.map(p => (
            <button
              key={p.id}
              className={`pattern-btn ${selectedPattern === p.id ? 'active' : ''}`}
              onClick={() => setSelectedPattern(selectedPattern === p.id ? null : p.id)}
              title={p.hint}
            >
              {p.label}
            </button>
          ))}
        </div>

        <div className="input-row">
          <input
            type="text"
            value={text}
            onChange={e => setText(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleSend()}
            placeholder={selectedPattern ? 'Напишите, как вы это скажете...' : 'Напишите ответ или выберите приём'}
          />
          <button className="send-btn" onClick={handleSend} disabled={!text.trim() && !selectedPattern}>
            Отправить
          </button>
        </div>
      </div>
    </div>
  );
}