import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { getMyArenaAssignments, getMyArenaHistory, getUserId } from '../../api';

export default function ArenaIndex() {
  const navigate = useNavigate();
  const [active, setActive] = useState([]);
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    async function load() {
      const userId = getUserId();
      if (!userId) {
        setError('Нет данных пользователя. Войдите заново.');
        setLoading(false);
        return;
      }

      try {
        const activeData = await getMyArenaAssignments(userId);
        setActive(activeData);
      } catch (err) {
        setError(err.message);
      }

      try {
        const historyData = await getMyArenaHistory(userId);
        setHistory(historyData);
      } catch {
        // история может не работать
      }

      setLoading(false);
    }
    load();
  }, []);

  function formatDate(iso) {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('ru-RU', {
      day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit'
    });
  }

  function statusLabel(status) {
    const map = {
      scheduled: 'Запланирована',
      ready: 'Готова к подключению',
      in_progress: 'Идёт',
      completed: 'Завершена',
      cancelled: 'Отменена',
      waiting_for_participants: 'Ожидание участников',
      reviewed: 'Проверена'
    };
    return map[status] || status;
  }

  if (loading) return <div className="page">Загрузка...</div>;

  return (
    <div className="page">
      <header>
        <h1>Арена переговоров</h1>
        <p>Переговоры с реальным партнёром. Без подсказок.</p>
      </header>

      <section className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/home')}>Обучение</button>
        <button onClick={() => navigate('/profile')}>Профиль</button>
      </section>

      {error && <p className="error">{error}</p>}

      <div className="card">
        <h2 className="section-title">Запланированные арены</h2>

        {active.length === 0 && !error && (
          <p className="muted">У вас пока нет назначенных арен. Куратор назначит — придёт уведомление.</p>
        )}

        {active.map(a => (
          <div key={a.arena_assignment_id} className="competency-item">
            <div style={{ flex: 1 }}>
              <strong>{a.scenario_title}</strong>
              <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>
                {formatDate(a.scheduled_at)}
              </p>
              <p className="muted" style={{ margin: '2px 0 0', fontSize: 12 }}>
                Статус: {statusLabel(a.status)}
              </p>
            </div>
            {(a.status === 'ready' || a.status === 'scheduled') && (
  <button onClick={() => navigate(`/arena/assignment/${a.arena_assignment_id}/join`)}>
    Подключиться
  </button>
)}
{(a.status === 'in_progress' || a.status === 'completed') && a.arena_session_id && (
  <button onClick={() => navigate(`/arena/session/${a.arena_session_id}`)}>
    {a.status === 'in_progress' ? 'Вернуться' : 'Смотреть'}
  </button>
)}
          </div>
        ))}
      </div>

      <div className="card">
        <h2 className="section-title">История арен</h2>

        {history.length === 0 && (
          <p className="muted">Пока нет завершённых арен.</p>
        )}

        <button onClick={() => navigate(`/arena/history/${h.arena_session_id}`)}>
  Смотреть
</button>
      </div>
    </div>
  );
}