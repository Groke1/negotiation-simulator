import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  joinArena,
  getArenaSession,
  getArenaMessages,
  completeArena,
  getToken,
  getMe
} from '../../api';

export default function ArenaSession() {
  const { assignmentId } = useParams();
  const navigate = useNavigate();

  const [session, setSession] = useState(null);
  const [messages, setMessages] = useState([]);
  const [me, setMe] = useState(null);
  const [text, setText] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [wsStatus, setWsStatus] = useState('connecting');
  const [completing, setCompleting] = useState(false);

  const wsRef = useRef(null);
  const scrollRef = useRef(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  useEffect(() => {
    let ws;

    async function load() {
      try {
        // 1. Join — получаем arena_session_id и websocket_path
        const joinData = await joinArena(assignmentId);
        const sessionId = joinData.arena_session_id;
        const wsPath = joinData.websocket_path;

        // 2. Загружаем сессию и себя
        const [sessionData, meData] = await Promise.all([
          getArenaSession(sessionId),
          getMe()
        ]);
        setSession(sessionData);
        setMe(meData);

        // 3. История сообщений
        try {
          const history = await getArenaMessages(sessionId);
          setMessages(history);
        } catch {}

        // 4. WebSocket
        const token = getToken();
        const wsUrl = `ws://localhost:8080${wsPath}`;
        ws = new WebSocket(wsUrl, ['arena-v1', token]);
        wsRef.current = ws;

        ws.onopen = () => setWsStatus('open');
        ws.onerror = () => setWsStatus('error');
        ws.onclose = () => setWsStatus('closed');

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);

            if (data.type === 'message.created') {
              setMessages(prev => [...prev, data.message]);
            } else if (data.type === 'participant.joined') {
              setMessages(prev => [...prev, {
                message_id: 'sys-' + Date.now(),
                sender_id: data.user_id,
                text: 'Подключился участник',
                created_at: new Date().toISOString(),
                _system: true
              }]);
            } else if (data.type === 'participant.left') {
              setMessages(prev => [...prev, {
                message_id: 'sys-' + Date.now(),
                sender_id: data.user_id,
                text: 'Участник вышел',
                created_at: new Date().toISOString(),
                _system: true
              }]);
            } else if (data.type === 'arena.completed') {
              setWsStatus('closed');
            } else if (data.type === 'error') {
              setError(data.message);
            }
          } catch (e) {
            console.error('WebSocket parse error:', e);
          }
        };
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    load();

    return () => {
      if (ws) ws.close();
    };
  }, [assignmentId]);

  function handleSend(e) {
    e.preventDefault();
    if (!text.trim()) return;
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      setError('WebSocket не подключён');
      return;
    }
    wsRef.current.send(JSON.stringify({
      type: 'message.send',
      message_id: crypto.randomUUID(),
      text: text.trim()
    }));
    setText('');
  }

  async function handleComplete() {
    if (!confirm('Завершить арену?')) return;
    setCompleting(true);
    try {
      await completeArena(session.arena_session_id);
      navigate('/arena');
    } catch (err) {
      setError(err.message);
    } finally {
      setCompleting(false);
    }
  }

  function formatTime(iso) {
    return new Date(iso).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  }

  if (loading) return <div className="page">Подключение...</div>;
  if (error && !session) return <div className="page"><p className="error">{error}</p></div>;

  return (
    <div className="chat-page arena-mode">
      <header className="chat-header">
        <button className="back-btn" onClick={() => navigate('/arena')}>←</button>
        <div className="chat-header-info">
          <h2>Арена · {session?.scenario_title}</h2>
          <p>
            {wsStatus === 'open' && 'подключено'}
            {wsStatus === 'connecting' && 'подключение...'}
            {wsStatus === 'closed' && 'закрыто'}
            {wsStatus === 'error' && 'ошибка'}
          </p>
        </div>
        {session?.status === 'in_progress' && (
          <button
            className="send-btn"
            style={{ marginLeft: 'auto', fontSize: 13, padding: '8px 14px' }}
            onClick={handleComplete}
            disabled={completing}
          >
            {completing ? '...' : 'Завершить'}
          </button>
        )}
      </header>

      <div className="chat-stream" ref={scrollRef}>
        {messages.length === 0 && (
          <p className="muted" style={{ textAlign: 'center', marginTop: 20 }}>
            Сообщений пока нет. Начните переговоры.
          </p>
        )}
        {messages.map(m => {
          const isMe = m.sender_id === me?.id;
          return (
            <div
              key={m.message_id}
              className={`msg ${isMe ? 'msg-user' : 'msg-partner'}`}
              style={m._system ? { alignSelf: 'center', opacity: 0.6 } : {}}
            >
              <span className="msg-sender">
                {m._system ? 'Система' : (isMe ? 'Вы' : 'Партнёр')}
              </span>
              <div className="msg-bubble">{m.text}</div>
              <span className="msg-pattern">{formatTime(m.created_at)}</span>
            </div>
          );
        })}
      </div>

      <div className="chat-input-panel">
        <form onSubmit={handleSend} className="input-row">
          <input
            value={text}
            onChange={e => setText(e.target.value)}
            placeholder={wsStatus === 'open' ? 'Напишите сообщение...' : 'Ожидание...'}
            disabled={wsStatus !== 'open'}
          />
          <button type="submit" className="send-btn" disabled={wsStatus !== 'open' || !text.trim()}>
            Отправить
          </button>
        </form>
        {error && <p className="error" style={{ marginTop: 8 }}>{error}</p>}
      </div>
    </div>
  );
}