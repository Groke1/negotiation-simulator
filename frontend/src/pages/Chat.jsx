import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { classifyPartnerMessage, getSemanticActions, getTrainingSession, sendTrainingAction, sendTrainingMessage } from '../api';
import { EmptyState, ErrorState, LoadingState, SemanticActionPicker } from '../components/UI';
import { runtimeLabel, semanticActionLabel } from '../app/presentation';
import { createId } from '../app/id';
import { sortMessagesChronologically } from '../app/messages';
import PublicScenarioViewer from '../components/PublicScenarioViewer';

const contentText = (message) => message?.content?.text || message?.content?.body || message?.content?.transcript || (message?.content?.media_id ? 'Голосовое сообщение' : '');
const time = (value) => value ? new Date(value).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }) : '';

export default function Chat() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const client = useQueryClient();
  const streamRef = useRef(null);
  const typingTimeoutRef = useRef(null);
  const [text, setText] = useState('');
  const [action, setAction] = useState('');
  const [partnerAction, setPartnerAction] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [typing, setTyping] = useState(false);
  const [feedback, setFeedback] = useState(null);
  const [error, setError] = useState('');
  const sessionQuery = useQuery({ queryKey: ['training-session', sessionId], queryFn: ({ signal }) => getTrainingSession(sessionId, signal) });
  const actionsQuery = useQuery({ queryKey: ['semantic-actions'], queryFn: ({ signal }) => getSemanticActions(signal), staleTime: 3600000 });
  const session = sessionQuery.data;
  const actions = actionsQuery.data;
  const availableActions = actions || [];
  const actionNames = useMemo(() => new Map((actions || []).map((item) => [item.code, semanticActionLabel(item)])), [actions]);
  const orderedMessages = useMemo(() => sortMessagesChronologically(session?.messages), [session?.messages]);
  const deterministic = session?.mode === 'deterministic';
  const canClassifyPartner = session?.mode === 'virtual_partner';
  const unclassified = canClassifyPartner ? orderedMessages.filter((message) => message.author === 'virtual_partner' && !message.classification).at(-1) : null;
  const currentIncoming = session?.current_node?.incoming_message;
  const currentIncomingInHistory = Boolean(currentIncoming) && orderedMessages.some((message) => message.author === 'virtual_partner' && (message.message_id === currentIncoming.message_id || contentText(message) === currentIncoming.text));
  const isFinished = session?.is_end || session?.status === 'completed';

  useEffect(() => {
    const stream = streamRef.current;
    if (stream) stream.scrollTop = stream.scrollHeight;
  }, [orderedMessages.length, currentIncoming?.message_id]);

  useEffect(() => () => window.clearTimeout(typingTimeoutRef.current), []);

  async function refresh() { await client.invalidateQueries({ queryKey: ['training-session', sessionId] }); }
  async function submitPartnerClassification() {
    if (!unclassified || !partnerAction) return;
    setSubmitting(true);
    setError('');
    try {
      const result = await classifyPartnerMessage(sessionId, unclassified.message_id, partnerAction);
      setFeedback({ ...result, messageId: unclassified.message_id, expectedName: actionNames.get(result.expected_semantic_action_code) });
      await refresh();
      setPartnerAction('');
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSubmitting(false);
    }
  }
  async function send() {
    if (!text.trim() || !action) return;
    setSubmitting(true);
    setError('');
    if (session?.mode === 'virtual_partner') setTyping(true);
    try {
      const result = await sendTrainingMessage(sessionId, { message_id: createId(), content: { type: 'chat', text: text.trim() }, semantic_action_code: action });
      setText('');
      setAction('');
      if (result.is_end) navigate(`/learner/training/${sessionId}/analysis`);
      else await refresh();
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSubmitting(false);
      if (session?.mode === 'virtual_partner') {
        window.clearTimeout(typingTimeoutRef.current);
        typingTimeoutRef.current = window.setTimeout(() => setTyping(false), 700);
      }
    }
  }
  async function chooseDeterministic(option) {
    if (!action) return;
    setSubmitting(true);
    setError('');
    try {
      const result = await sendTrainingAction(sessionId, { action_id: option.action_id, semantic_action_code: action, client_event_id: createId() });
      setFeedback({ ...result.classification, expectedName: actionNames.get(result.classification?.expected_semantic_action_code) });
      setAction('');
      if (result.is_end) navigate(`/learner/training/${sessionId}/analysis`);
      else await refresh();
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSubmitting(false);
    }
  }

  if (sessionQuery.isLoading || actionsQuery.isLoading) return <LoadingState />;
  if (sessionQuery.error || actionsQuery.error) return <ErrorState error={sessionQuery.error || actionsQuery.error} retry={() => { sessionQuery.refetch(); actionsQuery.refetch(); }} />;
  if (isFinished) return <div className="page"><EmptyState title="Тренировка завершена" description="Итог, история сообщений, выбранные семантические метки и изменения профиля доступны в отчёте." action={<button type="button" onClick={() => navigate(`/learner/training/${sessionId}/analysis`)}>Открыть отчёт</button>} /></div>;

  return <div className="chat-layout">
    <header className="chat-top"><div><span className="eyebrow">Тренировка · {runtimeLabel(session.mode)}</span><h2>{session.preparation?.title || session.scenario_title}</h2><p>{session.current_block ? `Текущее состояние: ${session.current_block.name}` : session.preparation?.partner_role?.name || 'Собеседник'}</p></div><PublicScenarioViewer scenario={session.scenario} preparation={session.preparation} currentBlock={session.current_block} /></header>
    <section className="chat-stream" ref={streamRef}>
      {orderedMessages.map((message) => <div key={message.message_id} className={`bubble ${message.author === 'learner' ? 'mine' : ''}`}><small>{message.author === 'learner' ? 'Вы' : 'Собеседник'}</small>{contentText(message)}<time className="message-time">{time(message.created_at)}</time>{message.author === 'learner' && message.semantic_action_code && <div className="message-semantic-report"><span>Выбранная метка:</span><strong>{actionNames.get(message.semantic_action_code) || message.semantic_action_code}</strong></div>}{message.classification && <div className="message-semantic-report partner-classification-result"><p><span>Ваш вариант:</span> <strong>{actionNames.get(message.classification.selected_semantic_action_code) || message.classification.selected_semantic_action_code}</strong></p><p className={message.classification.correct ? 'feedback good' : 'feedback bad'}>{message.classification.correct ? '✓ Действие распознано верно' : <>Не совпало. Правильная метка: <strong>{actionNames.get(message.classification.expected_semantic_action_code) || message.classification.expected_semantic_action_code}</strong></>}</p></div>}{message.message_id === unclassified?.message_id && <div className="opponent-classifier"><p className="opponent-classifier-hint">Определите действие собеседника, чтобы продолжить тренировку.</p><SemanticActionPicker actions={availableActions} value={partnerAction} onChange={setPartnerAction} disabled={submitting} result={feedback?.messageId === message.message_id ? feedback : null} label="Какое действие совершил собеседник?" /><button type="button" onClick={submitPartnerClassification} disabled={!partnerAction || submitting}>{submitting ? 'Проверяем…' : 'Проверить'}</button></div>}</div>)}
      {deterministic && currentIncoming && !currentIncomingInHistory && <div className="bubble"><small>Собеседник</small>{currentIncoming.text}</div>}
      {!deterministic && typing && <div className="bubble typing-bubble" aria-live="polite"><small>Собеседник</small><span>печатает<span className="typing-dots">…</span></span></div>}
      {error && <p className="error">{error}</p>}
    </section>
    <footer className="chat-compose">{deterministic ? <><SemanticActionPicker actions={availableActions} value={action} onChange={setAction} disabled={submitting} result={feedback} label="Какое действие вы совершаете?" /><div className="form-grid" style={{ marginTop: 12 }}>{session.current_node?.available_actions?.map((option) => <button className="secondary" type="button" key={option.action_id} disabled={!action || submitting} onClick={() => chooseDeterministic(option)}>{option.text}</button>)}</div></> : !unclassified && <><textarea rows="3" value={text} onChange={(event) => setText(event.target.value)} placeholder="Ваше сообщение…" disabled={submitting} /><footer><SemanticActionPicker actions={availableActions} value={action} onChange={setAction} disabled={submitting} label="Какое действие вы совершаете?" /><button type="button" onClick={send} disabled={!text.trim() || !action || submitting}>{submitting ? 'Отправляем…' : 'Отправить'}</button></footer></>}</footer>
  </div>;
}
