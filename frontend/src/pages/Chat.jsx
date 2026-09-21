import { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { mockScenario } from '../data/mockScenario';

export default function Chat() {
  const { scenarioId } = useParams();
  const navigate = useNavigate();

  const [scenario, setScenario] = useState(null);
  const [currentBlockId, setCurrentBlockId] = useState(null);
  const [currentNodeId, setCurrentNodeId] = useState(null);
  const [history, setHistory] = useState([]);
  const [isTyping, setIsTyping] = useState(false);
  const scrollRef = useRef(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [history]);

  useEffect(() => {
    // Берём сценарий из localStorage или из мока
    const saved = localStorage.getItem('scenario_' + scenarioId);
    let s;

    if (saved) {
      s = JSON.parse(saved);
    } else {
      s = { ...mockScenario, scenario_id: scenarioId };
    }

    setScenario(s);

    // Находим start-блок
    const startBlock = s.blocks?.find(b => b.type === 'start') || s.blocks?.[0];
    if (!startBlock) return;

    // Идём по переходу из старта
    const firstTransition = startBlock.transitions?.[0];
    if (!firstTransition) return;

    const firstBlock = s.blocks.find(b => b.id === firstTransition.target_block_id);
    if (!firstBlock) return;

    setCurrentBlockId(firstBlock.id);

    // Находим первый узел диалога
    const firstNodeId = firstBlock.deterministic_dialogue?.start_node_id
      || firstBlock.deterministic_dialogue?.nodes?.[0]?.id;
    setCurrentNodeId(firstNodeId);
  }, [scenarioId]);

  if (!scenario) return <div className="page">Загрузка...</div>;

  const currentBlock = scenario.blocks?.find(b => b.id === currentBlockId);
  const currentNode = currentBlock?.deterministic_dialogue?.nodes?.find(n => n.id === currentNodeId);

  function chooseReply(reply) {
    // Добавляем в историю
    const newHistory = [
      ...history,
      { partner: currentNode.incoming_message?.text || '', user: reply.text }
    ];
    setHistory(newHistory);

    setIsTyping(true);
    setTimeout(() => {
      // Если у варианта есть переход — идём в другой блок
      if (reply.transition_id) {
        const transition = currentBlock.transitions.find(t => t.id === reply.transition_id);
        if (transition) {
          const nextBlock = scenario.blocks.find(b => b.id === transition.target_block_id);
          if (nextBlock) {
            // Если следующий блок финальный — показываем финал
            if (nextBlock.type === 'end') {
              navigate(`/debrief/${scenario.scenario_id}/${nextBlock.id}`);
              return;
            }
            // Иначе идём в блок
            setCurrentBlockId(nextBlock.id);
            const firstNodeId = nextBlock.deterministic_dialogue?.start_node_id
              || nextBlock.deterministic_dialogue?.nodes?.[0]?.id;
            setCurrentNodeId(firstNodeId);
          }
        }
      } else if (reply.next_node_id) {
        // Идём к следующему узлу в этом же блоке
        setCurrentNodeId(reply.next_node_id);
      }
      setIsTyping(false);
    }, 800);
  }

  if (!currentNode) {
    return (
      <div className="page">
        <h1>Нет диалога в блоке</h1>
        <button onClick={() => navigate('/home')}>На главную</button>
      </div>
    );
  }

  return (
    <div className="chat-page">
      <header className="chat-header">
        <button className="back-btn" onClick={() => navigate('/home')}>←</button>
        <div className="chat-header-info">
          <h2>{scenario.title}</h2>
          <p>{scenario.roles?.[1]?.name || 'Партнёр'} · {scenario.difficulty}</p>
        </div>
      </header>

      <div className="chat-stream" ref={scrollRef}>
        {history.map((h, i) => (
          <div key={i} className="msg-group">
            {h.partner && (
              <div className="msg msg-partner">
                <span className="msg-sender">Партнёр</span>
                <div className="msg-bubble">{h.partner}</div>
              </div>
            )}
            <div className="msg msg-user">
              <span className="msg-sender">Вы</span>
              <div className="msg-bubble">{h.user}</div>
            </div>
          </div>
        ))}

        {currentNode.incoming_message && !isTyping && (
          <div className="msg msg-partner">
            <span className="msg-sender">Партнёр</span>
            <div className="msg-bubble">{currentNode.incoming_message.text}</div>
          </div>
        )}

        {isTyping && (
          <div className="msg msg-partner">
            <span className="msg-sender">Партнёр</span>
            <div className="msg-bubble typing">
              <span></span><span></span><span></span>
              <em>печатает</em>
            </div>
          </div>
        )}
      </div>

      <div className="chat-input-panel">
        <div className="patterns-row" style={{ flexDirection: 'column', gap: 8 }}>
          {(currentNode.reply_options || []).map(opt => (
            <button
              key={opt.id}
              className="pattern-btn"
              onClick={() => chooseReply(opt)}
              style={{ textAlign: 'left', padding: '10px 14px' }}
            >
              {opt.text}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}