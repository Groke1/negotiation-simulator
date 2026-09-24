import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { completeArena, getArenaMessages, getSemanticActions, getToken, joinArena } from '../../api';
import { webSocketUrl } from '../../api/client';
import { useAuth } from '../../app/AuthContext';
import { ErrorState, LoadingState, SemanticActionPicker } from '../../components/UI';
import { createId } from '../../app/id';

const mergeMessages = (current, incoming) => Array.from(new Map([...current, ...incoming].map((message) => [message.message_id, message])).values())
  .sort((left, right) => new Date(left.created_at) - new Date(right.created_at));
const messageTime = (value) => value ? new Date(value).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) : '';

export default function ArenaSession() {
  const { assignmentId } = useParams();
  const { me } = useAuth();
  const navigate = useNavigate();
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState('');
  const [action, setAction] = useState('');
  const [status, setStatus] = useState('Подключаемся…');
  const [error, setError] = useState('');
  const [finishing, setFinishing] = useState(false);
  const socket = useRef(null);
  const retry = useRef(null);
  const joining = useQuery({ queryKey: ['arena-join', assignmentId], queryFn: () => joinArena(assignmentId), retry: false });
  const actions = useQuery({ queryKey: ['semantic-actions'], queryFn: ({ signal }) => getSemanticActions(signal), staleTime: 60 * 60 * 1000 });
  const joined = joining.data;

  useEffect(() => {
    if (!joined?.arena_session_id) return undefined;
    let stopped = false;
    const connect = async () => {
      setStatus('Подключаемся…');
      try {
        const history = await getArenaMessages(joined.arena_session_id);
        if (!stopped) setMessages((current) => mergeMessages(current, history));
        const ws = new WebSocket(webSocketUrl(joined.websocket_path), ['arena-v1', getToken()]);
        socket.current = ws;
        ws.onopen = () => !stopped && setStatus('В сети');
        ws.onmessage = (event) => {
          const item = JSON.parse(event.data);
          if (item.type === 'message.created') setMessages((current) => mergeMessages(current, [item.message]));
          if (item.type === 'participant.joined') setStatus('Оба участника в сети');
          if (item.type === 'participant.left') setStatus('Собеседник временно отключился');
          if (item.type === 'arena.completed') navigate(`/learner/arenas/${joined.arena_session_id}`);
          if (item.type === 'error') setError(item.message || 'Сервер временно не принял сообщение.');
        };
        ws.onclose = () => {
          if (!stopped) {
            setStatus('Переподключаемся…');
            retry.current = window.setTimeout(connect, 2000);
          }
        };
        ws.onerror = () => setError('Не удалось подключить live-чат. Повторяем попытку.');
      } catch (requestError) {
        if (!stopped) {
          setError(requestError.message);
          retry.current = window.setTimeout(connect, 3000);
        }
      }
    };
    connect();
    return () => {
      stopped = true;
      window.clearTimeout(retry.current);
      socket.current?.close();
    };
  }, [joined, navigate]);

  function send() {
    if (!text.trim() || !action || socket.current?.readyState !== WebSocket.OPEN) return;
    socket.current.send(JSON.stringify({ type: 'message.send', message_id: createId(), text: text.trim(), semantic_action_code: action }));
    setText('');
    setAction('');
  }

  async function finish() {
    if (!joined?.arena_session_id) return;
    setFinishing(true);
    setError('');
    try {
      await completeArena(joined.arena_session_id);
      navigate(`/learner/arenas/${joined.arena_session_id}`);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setFinishing(false);
    }
  }

  if (joining.isLoading || actions.isLoading) return <LoadingState label="Подключаемся к Arena…" />;
  if (joining.error) return <ErrorState error={joining.error} retry={joining.refetch} />;

  const connected = status === 'В сети' || status === 'Оба участника в сети';
  return <div className="chat-layout">
    <header className="chat-top">
      <div><span className="eyebrow">Arena · живые переговоры</span><h2>{joined?.preparation?.title}</h2><p>{status}</p></div>
      <div className="header-actions"><button className="secondary" type="button" onClick={finish} disabled={finishing}>{finishing ? 'Завершаем…' : 'Завершить Arena'}</button><button className="secondary" type="button" onClick={() => navigate('/learner/arenas')}>Выйти</button></div>
    </header>
    <section className="chat-stream">
      {joined?.preparation && <div className="panel"><strong>{joined.preparation.my_role.name}</strong><p>{joined.preparation.expanded_context}</p></div>}
      {messages.map((message) => <div className={`bubble ${message.sender_id === me?.id ? 'mine' : ''}`} key={message.message_id}><small>{message.sender_id === me?.id ? 'Вы' : 'Собеседник'}</small>{message.text}<time className="message-time">{messageTime(message.created_at)}</time></div>)}
      {error && <p className="error">{error}</p>}
    </section>
    <footer className="chat-compose"><textarea rows="3" value={text} onChange={(event) => setText(event.target.value)} placeholder="Сообщение партнёру…" /><footer><SemanticActionPicker actions={actions.data} value={action} onChange={setAction} label="Характер сообщения" /><button type="button" onClick={send} disabled={!text.trim() || !action || !connected}>Отправить</button></footer></footer>
  </div>;
}
