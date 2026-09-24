import { useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createArenaFeedback,
  getArenaFeedback,
  getArenaMessages,
  getArenaSession
} from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../../components/UI';

export default function ArenaFeedback() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const client = useQueryClient();
  const [selectedParticipant, setSelectedParticipant] = useState('');
  const [comment, setComment] = useState('');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  const session = useQuery({ queryKey: ['arena-session', sessionId], queryFn: ({ signal }) => getArenaSession(sessionId, signal) });
  const messages = useQuery({ queryKey: ['arena-messages', sessionId], queryFn: ({ signal }) => getArenaMessages(sessionId, signal) });
  const feedback = useQuery({ queryKey: ['arena-feedback', sessionId], queryFn: ({ signal }) => getArenaFeedback(sessionId, signal) });

  const participants = session.data?.participants || [];
  const activeParticipant = selectedParticipant || participants[0]?.user_id || '';
  const feedbackByParticipant = useMemo(() => new Map((feedback.data || []).map((item) => [item.participant_id, item])), [feedback.data]);
  const currentFeedback = feedbackByParticipant.get(activeParticipant);

  async function submit(event) {
    event.preventDefault();
    setError('');
    if (!activeParticipant || !comment.trim()) {
      setError('Выберите участника и напишите комментарий.');
      return;
    }
    setSaving(true);
    try {
      await createArenaFeedback(sessionId, activeParticipant, comment.trim());
      setComment('');
      await client.invalidateQueries({ queryKey: ['arena-feedback', sessionId] });
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  if (session.isLoading || messages.isLoading || feedback.isLoading) return <LoadingState />;
  if (session.error || messages.error || feedback.error) return <ErrorState error={session.error || messages.error || feedback.error} retry={() => { session.refetch(); messages.refetch(); feedback.refetch(); }} />;

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <span className="eyebrow">Куратор · Arena</span>
          <h1>{session.data?.scenario_title || 'Фидбек по Arena'}</h1>
          <p>Transcript и разбор доступны после завершения Arena. Feedback оставляется отдельно каждому участнику.</p>
        </div>
        <div>
          <button className="secondary" type="button" onClick={() => navigate('/curator/arenas')}>К Arena</button>
        </div>
      </header>

      <div className="split">
        <section className="card">
          <h2>Диалог</h2>
          <StatusBadge value={session.data?.status} />
          {messages.data.length ? (
            <div className="chat-stream inline-chat">
              {messages.data.map((message) => (
                <div key={message.message_id} className="bubble">
                  <small>{message.sender_id}</small>
                  {message.text}
                </div>
              ))}
            </div>
          ) : <EmptyState title="Сообщений нет" />}
        </section>

        <section className="card">
          <h2>Feedback</h2>
          {participants.length ? (
            <div className="tabs">
              {participants.map((participant, index) => (
                <button key={participant.user_id} type="button" className={activeParticipant === participant.user_id ? 'active' : ''} onClick={() => setSelectedParticipant(participant.user_id)}>
                  Участник {index + 1}
                </button>
              ))}
            </div>
          ) : <EmptyState title="Участники не найдены" />}

          {currentFeedback ? (
            <article className="panel">
              <span className="eyebrow">Feedback уже отправлен</span>
              <p>{currentFeedback.comment || 'Комментарий не указан.'}</p>
              {currentFeedback.ratings?.length ? currentFeedback.ratings.map((rating) => <p key={rating.cluster_id}>{rating.cluster_id}: {rating.score}</p>) : <p>Оценки по кластерам не добавлялись.</p>}
              <small>{currentFeedback.created_at ? new Date(currentFeedback.created_at).toLocaleString('ru-RU') : ''}</small>
            </article>
          ) : (
            <form className="form-grid" onSubmit={submit}>
              <label className="field">
                Комментарий
                <textarea rows="6" value={comment} onChange={(event) => setComment(event.target.value)} placeholder="Что получилось, что можно улучшить..." />
              </label>
              {error && <p className="error">{error}</p>}
              <button type="submit" disabled={saving || !activeParticipant || !comment.trim()}>{saving ? 'Сохраняем…' : 'Отправить feedback'}</button>
            </form>
          )}
        </section>
      </div>
    </div>
  );
}
