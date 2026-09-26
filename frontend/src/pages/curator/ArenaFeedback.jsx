import { useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createArenaFeedback,
  getCompetencyClusters,
  getArenaFeedback,
  getArenaMessages,
  getArenaSession
} from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge } from '../../components/UI';
import { formatDelta, formatNumber } from '../../app/format';
import { sortMessagesChronologically } from '../../app/messages';

export default function ArenaFeedback() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const client = useQueryClient();
  const [selectedParticipant, setSelectedParticipant] = useState('');
  const [comment, setComment] = useState('');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);
  const [ratings, setRatings] = useState([]);

  const session = useQuery({ queryKey: ['arena-session', sessionId], queryFn: ({ signal }) => getArenaSession(sessionId, signal) });
  const messages = useQuery({ queryKey: ['arena-messages', sessionId], queryFn: ({ signal }) => getArenaMessages(sessionId, signal) });
  const feedback = useQuery({ queryKey: ['arena-feedback', sessionId], queryFn: ({ signal }) => getArenaFeedback(sessionId, signal) });
  const clusters = useQuery({ queryKey: ['competency-clusters'], queryFn: ({ signal }) => getCompetencyClusters(signal), staleTime: 3600000 });

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
      await createArenaFeedback(sessionId, activeParticipant, comment.trim(), ratings.filter((rating) => rating.score !== ''));
      setComment('');
      await client.invalidateQueries({ queryKey: ['arena-feedback', sessionId] });
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  if (session.isLoading || messages.isLoading || feedback.isLoading || clusters.isLoading) return <LoadingState />;
  if (session.error || messages.error || feedback.error || clusters.error) return <ErrorState error={session.error || messages.error || feedback.error || clusters.error} retry={() => { session.refetch(); messages.refetch(); feedback.refetch(); clusters.refetch(); }} />;

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
              {sortMessagesChronologically(messages.data).map((message) => (
                <div key={message.message_id} className="bubble">
                  <small>{message.sender_id === activeParticipant ? 'Выбранный участник' : 'Другой участник'}</small>
                  {message.text}
                  <time className="message-time">{message.created_at ? new Date(message.created_at).toLocaleString('ru-RU') : ''}</time>
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
              {currentFeedback.ratings?.length ? currentFeedback.ratings.map((rating) => <p key={rating.cluster_id}>{clusters.data.find((cluster) => cluster.id === rating.cluster_id)?.name || 'Компетенция'}: {rating.score}</p>) : <p>Оценки по кластерам не добавлялись.</p>}
              {currentFeedback.profile_change?.length ? <section><h3>Изменение профиля</h3>{currentFeedback.profile_change.map((change) => <p key={change.cluster_id}>{clusters.data.find((cluster) => cluster.id === change.cluster_id)?.name || 'Компетенция'}: {formatNumber(change.before)} → {formatNumber(change.after)} ({formatDelta(change.delta)})</p>)}</section> : <p>Backend не вернул изменений профиля для этого feedback.</p>}
              <small>{currentFeedback.created_at ? new Date(currentFeedback.created_at).toLocaleString('ru-RU') : ''}</small>
            </article>
          ) : session.data?.status !== 'completed' ? (
            <EmptyState title="Feedback станет доступен после завершения Arena" description="Сейчас можно просмотреть состав участников и ход диалога." />
          ) : (
            <form className="form-grid" onSubmit={submit}>
              <label className="field">
                Комментарий
                <textarea rows="6" value={comment} onChange={(event) => setComment(event.target.value)} placeholder="Что получилось, что можно улучшить..." />
              </label>
              <fieldset className="choice-group"><legend>Оценки компетенций (необязательно)</legend>{clusters.data.map((cluster) => { const rating = ratings.find((item) => item.cluster_id === cluster.id); return <label className="field" key={cluster.id}>{cluster.name}<input type="number" min="0" max="100" value={rating?.score ?? ''} placeholder="0–100" onChange={(event) => setRatings((current) => { const next = current.filter((item) => item.cluster_id !== cluster.id); return event.target.value === '' ? next : [...next, { cluster_id: cluster.id, score: Number(event.target.value) }]; })} /></label>; })}</fieldset>
              {error && <p className="error">{error}</p>}
              <button type="submit" disabled={saving || !activeParticipant || !comment.trim()}>{saving ? 'Сохраняем…' : 'Отправить feedback'}</button>
            </form>
          )}
        </section>
      </div>
    </div>
  );
}
