import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { classifyArenaMessage, completeArena, getArenaMessages, getArenaPresence, getOrganizationUsers, getSemanticActions, getToken, joinArena, sendArenaMessage, updateArenaTyping } from '../../api';
import { webSocketUrl } from '../../api/client';
import { useAuth } from '../../app/AuthContext';
import { resolveActiveRole } from '../../app/roles';
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
  const [presence, setPresence] = useState([]);
  const [typingOpponent, setTypingOpponent] = useState(null);
  const streamRef = useRef(null);
  const socket = useRef(null);
  const retry = useRef(null);
  const connectionTimeout = useRef(null);
  const shouldReconnect = useRef(true);
  const typingTimeout = useRef(null);
  const typingActive = useRef(false);
  const joining = useQuery({ queryKey: ['arena-join', assignmentId], queryFn: () => joinArena(assignmentId), retry: false });
  const actions = useQuery({ queryKey: ['semantic-actions'], queryFn: ({ signal }) => getSemanticActions(signal), staleTime: 60 * 60 * 1000 });
  const organizationUsers = useQuery({ queryKey: ['organization-users', 'arena-chat'], queryFn: ({ signal }) => getOrganizationUsers({}, signal), staleTime: 5 * 60 * 1000, enabled: resolveActiveRole(me) === 'curator' });
  const joined = joining.data;
  const actionNames = useMemo(() => new Map((actions.data || []).map((item) => [item.code, semanticActionLabel(item)])), [actions.data]);
  const participantNames = useMemo(() => new Map([...(organizationUsers.data || []).map((user) => [user.id, user.username || user.name || user.email]), ...presence.map((participant) => [participant.user_id, participant.username]), ...(me?.id ? [[me.id, me.username || me.email || 'Вы']] : [])]), [me, organizationUsers.data, presence]);
  const unclassified = messages.filter((message) => message.sender_id !== me?.id && !message.classification).at(-1);

  const setParticipantPresence = (participant) => {
    if (!participant?.user_id) return;
    setPresence((current) => {
      const exists = current.some((item) => item.user_id === participant.user_id);
      return exists ? current.map((item) => item.user_id === participant.user_id ? { ...item, ...participant } : item) : [...current, participant];
    });
  };

  const sendTyping = (isTyping) => {
    if (!joined?.arena_session_id || typingActive.current === isTyping) return;
    typingActive.current = isTyping;
    const payload = JSON.stringify({ type: 'typing.update', is_typing: isTyping });
    if (socket.current?.readyState === WebSocket.OPEN) socket.current.send(payload);
    else updateArenaTyping(joined.arena_session_id, isTyping).catch(() => {
      // Presence notifications are supplementary: message delivery must remain usable after a transient typing error.
    });
  };

  const stopTyping = () => {
    window.clearTimeout(typingTimeout.current);
    sendTyping(false);
  };

  useEffect(() => {
    const stream = streamRef.current;
    if (stream) stream.scrollTop = stream.scrollHeight;
  }, [messages.length]);

  useEffect(() => {
    if (!joined?.arena_session_id) return undefined;
    let stopped = false;
    let reconnectAttempt = 0;
    shouldReconnect.current = true;

    const scheduleReconnect = () => {
      if (stopped || !shouldReconnect.current) return;
      reconnectAttempt += 1;
      const delay = Math.min(1000 * (2 ** Math.min(reconnectAttempt, 3)), 8000);
      setStatus('Переподключаемся…');
      retry.current = window.setTimeout(connect, delay);
    };

    const hydrate = async () => {
      try {
        const [history, presenceSnapshot] = await Promise.all([getArenaMessages(joined.arena_session_id), getArenaPresence(joined.arena_session_id)]);
        if (!stopped) {
          setMessages((current) => mergeMessages(current, history));
          setPresence(presenceSnapshot?.participants || []);
        }
      } catch {
        // The live channel must still be attempted when history or presence is temporarily unavailable.
      }
    };

    const connect = () => {
      if (stopped || !shouldReconnect.current) return;
      const token = getToken();
      if (!token) {
        setStatus('Не подключено');
        setError('Не удалось подтвердить сессию. Обновите страницу и войдите снова.');
        return;
      }
      setStatus(reconnectAttempt ? 'Переподключаемся…' : 'Подключаемся…');
      let ws;
      try {
        ws = new WebSocket(webSocketUrl(joined.websocket_path, { access_token: token }), 'arena-v1');
      } catch {
        setStatus('Подключение не установлено');
        setError('Не удалось открыть live-чат. Повторяем попытку.');
        scheduleReconnect();
        return;
      }
      socket.current = ws;
      connectionTimeout.current = window.setTimeout(() => {
        if (ws.readyState === WebSocket.CONNECTING) {
          setStatus('Подключение не установлено');
          setError('Не удалось подключиться к live-чату. Повторяем попытку.');
          ws.close();
        }
      }, 10000);
      ws.onopen = () => {
        window.clearTimeout(connectionTimeout.current);
        reconnectAttempt = 0;
        setError('');
        setStatus('В сети');
      };
      ws.onmessage = (event) => {
        let item;
        try { item = JSON.parse(event.data); } catch { return; }
        if (item.type === 'message.created') setMessages((current) => mergeMessages(current, [item.message]));
        if (item.type === 'presence.snapshot') setPresence(item.participants || []);
        if (item.type === 'participant.online' || item.type === 'participant.offline') {
          setParticipantPresence(item.participant);
          if (item.type === 'participant.offline' && item.participant?.user_id !== me?.id) setTypingOpponent(null);
        }
        if (item.type === 'typing.updated' && item.user_id !== me?.id) setTypingOpponent(item.is_typing ? item : null);
        if (item.type === 'arena.completed') {
          shouldReconnect.current = false;
          setCompletedBy(item.completed_by || 'other');
          setStatus(item.completed_by === me?.id ? 'Arena завершена вами' : 'Arena завершена другим участником');
          ws.close();
          navigate(`/learner/arenas/${joined.arena_session_id}`);
        }
        if (item.type === 'error') setError(item.code === 'ARENA_WAITING_FOR_PARTICIPANT' ? 'Ожидаем подключения второго участника. Отправка сообщений станет доступна, когда он будет в сети.' : item.code === 'INVALID_ARENA_STATE' ? 'Сейчас Arena не позволяет отправить сообщение. Проверьте, подключился ли собеседник.' : item.message || 'Сервер временно не принял сообщение.');
      };
      ws.onclose = () => {
        window.clearTimeout(connectionTimeout.current);
        if (socket.current === ws) socket.current = null;
        scheduleReconnect();
      };
      ws.onerror = () => {
        setStatus('Подключение не установлено');
      };
    };
    void hydrate();
    connect();
    return () => {
      stopped = true;
      shouldReconnect.current = false;
      window.clearTimeout(retry.current);
      window.clearTimeout(connectionTimeout.current);
      window.clearTimeout(typingTimeout.current);
      socket.current?.close();
    };
  }, [joined, me?.id, navigate]);

  async function send() {
    const opponent = presence.find((participant) => participant.user_id !== me?.id);
    if (!text.trim() || !action || completedBy) return;
    if (!opponent?.online) {
      setError('Ожидаем подключения второго участника. Отправка сообщений станет доступна, когда он будет в сети.');
      return;
    }
    const message = { message_id: createId(), content: { type: 'chat', text: text.trim() }, semantic_action_code: action };
    stopTyping();
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

  function handleTextChange(event) {
    const nextText = event.target.value;
    setText(nextText);
    const shouldNotify = Boolean(nextText.trim());
    if (shouldNotify !== typingActive.current) sendTyping(shouldNotify);
    window.clearTimeout(typingTimeout.current);
    if (shouldNotify) typingTimeout.current = window.setTimeout(stopTyping, 1800);
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

  const opponent = presence.find((participant) => participant.user_id !== me?.id);
  const opponentOnline = Boolean(opponent?.online);
  const opponentName = opponent?.username || 'собеседника';
  return <div className="chat-layout">
    <header className="chat-top">
      <div><span className="eyebrow">Arena · живые переговоры</span><h2>{joined?.scenario?.title || joined?.preparation?.title}</h2><p className={`arena-presence ${opponentOnline ? 'online' : 'offline'}`}><i aria-hidden="true" />{opponent ? `${opponentName} ${opponentOnline ? 'в сети' : 'не в сети'}` : 'Ожидаем подключения собеседника'} <span>· {status}</span></p></div>
      <div className="header-actions"><PublicScenarioViewer scenario={joined?.scenario} preparation={joined?.preparation} /><button className="secondary" type="button" onClick={finish} disabled={finishing}>{finishing ? 'Завершаем…' : 'Завершить Arena'}</button><button className="secondary" type="button" onClick={() => navigate('/learner/arenas')}>Выйти</button></div>
    </header>
    <section className="chat-stream" ref={streamRef}>
      {messages.map((message) => { const mine = message.sender_id === me?.id; const pendingClassification = message.message_id === unclassified?.message_id; const classifying = classifyingMessageId === message.message_id; return <div className={`bubble ${mine ? 'mine' : ''}`} key={message.message_id}><small>{participantNames.get(message.sender_id) || (mine ? (me?.username || me?.email || 'Вы') : 'Собеседник')}</small><div className="message-content">{messageText(message)}</div><time className="message-time">{messageTime(message.created_at)}</time><ArenaMessageSemanticReport message={message} mine={mine} actionNames={actionNames} />{pendingClassification && <div className="opponent-classifier"><p className="opponent-classifier-hint">Определите действие собеседника, чтобы продолжить переговоры.</p><SemanticActionPicker actions={actions.data || []} value={partnerAction} onChange={setPartnerAction} disabled={Boolean(classifyingMessageId)} label="Какое действие совершил собеседник?" /><button type="button" onClick={() => classifyOpponentMessage(message)} disabled={!partnerAction || Boolean(classifyingMessageId)}>{classifying ? 'Проверяем…' : 'Проверить'}</button></div>}</div>; })}
      {typingOpponent && <div className="typing-bubble arena-typing">{typingOpponent.username || opponentName} печатает<span className="typing-dots">...</span></div>}
      {error && <p className="error">{error}</p>}
    </section>
    <footer className="chat-compose">{!completedBy && !unclassified && <>{!opponentOnline && <p className="arena-composer-notice">{opponent ? `${opponentName} сейчас не в сети. Отправка станет доступна после подключения.` : 'Ожидаем подключения второго участника. Отправка станет доступна, когда он будет в сети.'}</p>}<textarea rows="3" value={text} onChange={handleTextChange} onBlur={stopTyping} placeholder="Сообщение партнёру…" disabled={!opponentOnline} /><footer><SemanticActionPicker actions={actions.data || []} value={action} onChange={setAction} label="Какое действие вы совершаете?" disabled={!opponentOnline} /><button type="button" onClick={send} disabled={!text.trim() || !action || !opponentOnline}>Отправить</button></footer></>}</footer>
  </div>;
}
