import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  getArenaSession,
  getArenaMessages,
  getArenaFeedback,
  createArenaFeedback
} from '../../api';

export default function ArenaFeedback() {
  const { sessionId } = useParams();
  const navigate = useNavigate();

  const [session, setSession] = useState(null);
  const [messages, setMessages] = useState([]);
  const [feedback, setFeedback] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [selectedParticipant, setSelectedParticipant] = useState('');
  const [comment, setComment] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const sessionData = await getArenaSession(sessionId);
        setSession(sessionData);

        try {
          const messagesData = await getArenaMessages(sessionId);
          setMessages(messagesData);
        } catch {}

        try {
          const feedbackData = await getArenaFeedback(sessionId);
          setFeedback(feedbackData);
        } catch {}
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [sessionId]);

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!selectedParticipant || !comment.trim()) {
      setError('Выберите участника и напишите комментарий');
      return;
    }

    setSaving(true);
    try {
      await createArenaFeedback(sessionId, selectedParticipant, comment.trim());
      setSuccess('Фидбек сохранён');
      setComment('');

      // Обновляем список фидбеков
      const feedbackData = await getArenaFeedback(sessionId);
      setFeedback(feedbackData);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  function formatTime(iso) {
    return new Date(iso).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  }

  if (loading) return <div className="page">Загрузка...</div>;
  if (error && !session) return <div className="page"><p className="error">{error}</p></div>;

  return (
    <div className="page">
      <header>
        <h1>Фидбек по арене</h1>
        <p>{session?.scenario_title}</p>
      </header>

      <section className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/curator')}>Назад к команде</button>
        <button onClick={() => navigate('/curator/arena')}>Назначить арену</button>
      </section>

      <div className="card">
        <h2 className="section-title">Диалог</h2>

        {messages.length === 0 && (
          <p className="muted">Сообщений нет.</p>
        )}

        {messages.map(m => (
          <div key={m.message_id} className="competency-item" style={{ flexDirection: 'column', alignItems: 'flex-start' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
              <strong style={{ fontSize: 13 }}>Участник</strong>
              <span className="muted" style={{ fontSize: 12 }}>{formatTime(m.created_at)}</span>
            </div>
            <p style={{ margin: '6px 0 0', fontSize: 14 }}>{m.text}</p>
          </div>
        ))}
      </div>

      <div className="card">
        <h2 className="section-title">Оставить фидбек</h2>

        <form onSubmit={handleSubmit}>
          <label className="field">
            <span>Кому</span>
            <select value={selectedParticipant} onChange={e => setSelectedParticipant(e.target.value)} required>
              <option value="">— выберите участника —</option>
              {session?.participants?.map(p => (
                <option key={p.user_id} value={p.user_id}>
                  {p.role === 'learner' ? 'Ученик' : 'Партнёр'} — {p.user_id.slice(0, 8)}
                </option>
              ))}
            </select>
          </label>

          <label className="field">
            <span>Комментарий</span>
            <textarea
              value={comment}
              onChange={e => setComment(e.target.value)}
              rows={5}
              placeholder="Что получилось, что можно улучшить..."
              required
            />
          </label>

          {error && <p className="error">{error}</p>}
          {success && <p style={{ color: '#0a7c2f', fontSize: 14, fontWeight: 600 }}>{success}</p>}

          <button type="submit" disabled={saving}>
            {saving ? 'Сохранение...' : 'Сохранить фидбек'}
          </button>
        </form>
      </div>

      {feedback.length > 0 && (
        <div className="card">
          <h2 className="section-title">Уже оставленные фидбеки</h2>
          {feedback.map(f => (
            <div key={f.feedback_id} className="competency-item" style={{ flexDirection: 'column', alignItems: 'flex-start' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                <strong style={{ fontSize: 13 }}>Участник: {f.participant_id.slice(0, 8)}</strong>
                <span className="muted" style={{ fontSize: 12 }}>
                  {new Date(f.created_at).toLocaleDateString('ru-RU')}
                </span>
              </div>
              <p style={{ margin: '6px 0 0', fontSize: 14 }}>{f.comment}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}