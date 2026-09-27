import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { classifyArenaMessage, completeArena, getArenaMessages, getOrganizationUsers, getSemanticActions, getToken, joinArena, sendArenaMessage } from '../../api';
import { webSocketUrl } from '../../api/client';
import { useAuth } from '../../app/AuthContext';
import { ErrorState, LoadingState, SemanticActionPicker } from '../../components/UI';
import { createId } from '../../app/id';
import { sortMessagesChronologically } from '../../app/messages';
import { semanticActionLabel } from '../../app/presentation';
import PublicScenarioViewer from '../../components/PublicScenarioViewer';
import ArenaMessageSemanticReport from '../../components/ArenaMessageSemanticReport';

const mergeMessages = (current, incoming) => sortMessagesChronologically(Array.from(new Map([...current, ...incoming].map((message) => [message.message_id, message])).values()));
const messageTime = (value) => value ? new Date(value).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) : '';
const messageText = (message) => {
  const content = message?.content || {};
  if (content.subject && content.body) return `${content.subject}\n\n${content.body}`;
  return content.text || content.body || content.transcript || (content.media_id ? 'Голосовое сообщение' : message?.text || '');
};

export default function ArenaSession() {
  const { assignmentId } = useParams();
  const { me } = useAuth();
  const navigate = useNavigate();
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState('');
  const [action, setAction] = useState('');
  const [partnerAction, setPartnerAction] = useState('');
  const [classifyingMessageId, setClassifyingMessageId] = useState('');
  const [status, setStatus] = useState('Подключаемся…');
  const [error, setError] = useState('');
  const [finishing, setFinishing] = useState(false);
  const [completedBy, setCompletedBy] = useState('');
  const streamRef = useRef(null);
  const socket = useRef(null);
  const retry = useRef(null);
  const joining = useQuery({ queryKey: ['arena-join', assignmentId], queryFn: () => joinArena(assignmentId), retry: false });
  const actions = useQuery({ queryKey: ['semantic-actions'], queryFn: ({ signal }) => getSemanticActions(signal), staleTime: 60 * 60 * 1000 });
  const organizationUsers = useQuery({ queryKey: ['organization-users', 'arena-chat'], queryFn: ({ signal }) => getOrganizationUsers({}, signal), staleTime: 5 * 60 * 1000 });
  const joined = joining.data;
  const actionNames = useMemo(() => new Map((actions.data || []).map((item) => [item.code, semanticActionLabel(item)])), [actions.data]);
  const participantNames = useMemo(() => new Map((organizationUsers.data || []).map((user) => [user.id, user.username || user.name || user.email])), [organizationUsers.data]);
  const unclassified = messages.filter((message) => message.sender_id !== me?.id && !message.classification).at(-1);

  useEffect(() => {
    const stream = streamRef.current;
    if (stream) stream.scrollTop = stream.scrollHeight;
  }, [messages.length]);

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
          if (item.type === 'arena.completed') {
            setCompletedBy(item.completed_by || 'other');
            setStatus(item.completed_by === me?.id ? 'Arena завершена вами' : 'Arena завершена другим участником');
            ws.close();
            navigate(`/learner/arenas/${joined.arena_session_id}`);
          }
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
  }, [joined, me?.id, navigate]);

  async function send() {
    if (!text.trim() || !action || completedBy) return;
    const message = { message_id: createId(), content: { type: 'chat', text: text.trim() }, semantic_action_code: action };
    if (socket.current?.readyState === WebSocket.OPEN) socket.current.send(JSON.stringify({ type: 'message.send', ...message }));
    else if (joined?.arena_session_id) {
      setError('');
      try {
        const saved = await sendArenaMessage(joined.arena_session_id, message);
        setMessages((current) => mergeMessages(current, [saved]));
      } catch (requestError) { setError(requestError.message); return; }
    } else return;
    setText('');
    setAction('');
  }

  async function classifyOpponentMessage(message) {
    if (!message || !partnerAction || !joined?.arena_session_id) return;
    setClassifyingMessageId(message.message_id);
    setError('');
    try {
      const classification = await classifyArenaMessage(joined.arena_session_id, message.message_id, { semantic_action_code: partnerAction });
      setMessages((current) => current.map((item) => item.message_id === message.message_id ? { ...item, classification } : item));
      setPartnerAction('');
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setClassifyingMessageId('');
    }
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
      <div><span className="eyebrow">Arena · живые переговоры</span><h2>{joined?.scenario?.title || joined?.preparation?.title}</h2><p>{status}</p></div>
      <div className="header-actions"><PublicScenarioViewer scenario={joined?.scenario} preparation={joined?.preparation} /><button className="secondary" type="button" onClick={finish} disabled={finishing}>{finishing ? 'Завершаем…' : 'Завершить Arena'}</button><button className="secondary" type="button" onClick={() => navigate('/learner/arenas')}>Выйти</button></div>
    </header>
    <section className="chat-stream" ref={streamRef}>
      {joined?.preparation && <div className="panel"><strong>{joined.preparation.my_role.name}</strong><p>{joined.preparation.expanded_context}</p></div>}
      {messages.map((message) => { const mine = message.sender_id === me?.id; const pendingClassification = message.message_id === unclassified?.message_id; const classifying = classifyingMessageId === message.message_id; return <div className={`bubble ${mine ? 'mine' : ''}`} key={message.message_id}><small>{participantNames.get(message.sender_id) || (mine ? (me?.username || me?.email || 'Вы') : 'Собеседник')}</small><div className="message-content">{messageText(message)}</div><time className="message-time">{messageTime(message.created_at)}</time><ArenaMessageSemanticReport message={message} mine={mine} actionNames={actionNames} />{pendingClassification && <div className="opponent-classifier"><p className="opponent-classifier-hint">Определите действие собеседника, чтобы продолжить переговоры.</p><SemanticActionPicker actions={actions.data || []} value={partnerAction} onChange={setPartnerAction} disabled={Boolean(classifyingMessageId)} label="Какое действие совершил собеседник?" /><button type="button" onClick={() => classifyOpponentMessage(message)} disabled={!partnerAction || Boolean(classifyingMessageId)}>{classifying ? 'Проверяем…' : 'Проверить'}</button></div>}</div>; })}
      {error && <p className="error">{error}</p>}
    </section>
    <footer className="chat-compose">{!completedBy && !unclassified && <><textarea rows="3" value={text} onChange={(event) => setText(event.target.value)} placeholder="Сообщение партнёру…" /><footer><SemanticActionPicker actions={actions.data || []} value={action} onChange={setAction} label="Какое действие вы совершаете?" /><button type="button" onClick={send} disabled={!text.trim() || !action || (!connected && !joined?.arena_session_id)}>Отправить</button></footer></>}</footer>
  </div>;
}
