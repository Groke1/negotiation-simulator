import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getArenaHistoryDetail, getUserId } from '../../api';

export default function ArenaHistory() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      const userId = getUserId();
      if (!userId) {
        setError('Нет данных пользователя');
        setLoading(false);
        return;
      }

      try {
        const historyData = await getArenaHistoryDetail(userId, sessionId);
        setData(historyData);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [sessionId]);

  function formatTime(iso) {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('ru-RU', {
      day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit'
    });
  }

  if (loading) return <div className="page">Загрузка...</div>;
  if (error) return <div className="page"><p className="error">{error}</p></div>;
  if (!data) return <div className="page">Нет данных</div>;

  const session = data.session || {};
  const messages = data.messages || [];
  const feedback = data.feedback || [];

  return (
    <div className="page">
      <header>
        <h1>История арены</h1>
        <p>{session.scenario_title}</p>
      </header>

      <section className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/arena')}>Назад к аренам</button>
      </section>

      <div className="card">
        <h2 className="section-title">Метаданные</h2>
        <div className="competency-item">
          <span>Статус</span>
          <strong>{session.status}</strong>
        </div>
        <div className="competency-item">
          <span>Запланирована</span>
          <strong>{formatTime(session.scheduled_at)}</strong>
        </div>
        <div className="competency-item">
          <span>Начата</span>
          <strong>{formatTime(session.started_at)}</strong>
        </div>
        <div className="competency-item">
          <span>Завершена</span>
          <strong>{formatTime(session.completed_at)}</strong>
        </div>
      </div>

      <div className="card">
        <h2 className="section-title">Сообщения ({messages.length})</h2>
        {messages.length === 0 && (
          <p className="muted">Сообщений не было.</p>
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

      {feedback.length > 0 && (
        <div className="card">
          <h2 className="section-title">Фидбек куратора</h2>
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