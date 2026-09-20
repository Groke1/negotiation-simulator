import { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario } from '../../data/scenario';
import { detectPattern } from '../../data/patterns';

export default function ArenaChat() {
  const { scenarioId } = useParams();
  const navigate = useNavigate();
  const scenario = getScenario(scenarioId);

  const [currentNodeId, setCurrentNodeId] = useState('n1');
  const [history, setHistory] = useState([]);
  const [isTyping, setIsTyping] = useState(false);
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
        <button onClick={() => navigate('/arena')}>Назад</button>
      </div>
    );
  }

  const node = scenario.nodes.find(n => n.id === currentNodeId);

  function handleSend() {
    if (!text.trim()) return;

    const detected = detectPattern(text);
    const action = node.actions.find(a => a.pattern === detected) || node.actions[0];

    const newHistory = [
      ...history,
      { partner: node.text, user: text, pattern: action.pattern }
    ];
    setHistory(newHistory);
    setUsedPatterns([...usedPatterns, action.pattern]);
    setText('');

    if (action.next.startsWith('end_')) {
      const notes = JSON.parse(localStorage.getItem('arena_notes') || '{}');
      localStorage.setItem('arena_result', JSON.stringify({
        scenarioId: scenario.id,
        outcome: action.next,
        patterns: [...usedPatterns, action.pattern],
        notes
      }));
      navigate(`/arena/debrief/${scenario.id}/${action.next}`);
      return;
    }

    setIsTyping(true);
    setTimeout(() => {
      setCurrentNodeId(action.next);
      setIsTyping(false);
    }, 1500);
  }

  return (
    <div className="chat-page arena-mode">
      <header className="chat-header">
        <button className="back-btn" onClick={() => navigate('/arena')}>←</button>
        <div className="chat-header-info">
          <h2>Арена · {scenario.title}</h2>
          <p>{scenario.roles.learner} · {scenario.roles.partner}</p>
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
        <div className="input-row">
          <input
            type="text"
            value={text}
            onChange={e => setText(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleSend()}
            placeholder="Напишите ответ..."
          />
          <button className="send-btn" onClick={handleSend} disabled={!text.trim()}>
            Отправить
          </button>
        </div>
      </div>
    </div>
  );
}