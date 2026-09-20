import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  getCuratorTeam,
  getScenarios,
  createArenaAssignment,
  getCuratorArenas
} from '../../api';

export default function CuratorArena() {
  const navigate = useNavigate();
  const [team, setTeam] = useState([]);
  const [scenarios, setScenarios] = useState([]);
  const [arenas, setArenas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [saving, setSaving] = useState(false);

  const [scenarioId, setScenarioId] = useState('');
  const [learnerA, setLearnerA] = useState('');
  const [learnerB, setLearnerB] = useState('');
  const [scheduledAt, setScheduledAt] = useState('');

  async function loadAll() {
    try {
      const teamData = await getCuratorTeam();
      setTeam(teamData);
    } catch (err) {
      setError(err.message);
    }

    try {
      const scenariosData = await getScenarios();
      setScenarios(scenariosData.filter(s => s.status === 'published'));
    } catch {}

    try {
      const arenasData = await getCuratorArenas();
      setArenas(arenasData);
    } catch {}
  }

  useEffect(() => {
    loadAll().finally(() => setLoading(false));
  }, []);

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (learnerA === learnerB) {
      setError('Нельзя выбрать одного и того же ученика');
      return;
    }

    setSaving(true);
    try {
      await createArenaAssignment(
        scenarioId,
        [
          { user_id: learnerA, role: 'learner' },
          { user_id: learnerB, role: 'partner' }
        ],
        new Date(scheduledAt).toISOString()
      );
      setSuccess('Арена назначена');
      setScenarioId('');
      setLearnerA('');
      setLearnerB('');
      setScheduledAt('');
      await loadAll();
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  function formatDate(iso) {
    if (!iso) return '—';
    return new Date(iso).toLocaleString('ru-RU', {
      day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit'
    });
  }

  function statusLabel(status) {
    const map = {
      scheduled: 'Запланирована',
      ready: 'Готова',
      in_progress: 'Идёт',
      completed: 'Завершена',
      cancelled: 'Отменена'
    };
    return map[status] || status;
  }

  if (loading) return <div className="page">Загрузка...</div>;

  return (
    <div className="page">
      <header>
        <h1>Назначить арену</h1>
        <p>Выберите сценарий, двух учеников и дату переговоров.</p>
      </header>

      <section className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/curator')}>Назад к команде</button>
      </section>

      <form onSubmit={handleSubmit} className="card">
        <h2 className="section-title">Новая арена</h2>

        <label className="field">
          <span>Сценарий</span>
          <select value={scenarioId} onChange={e => setScenarioId(e.target.value)} required>
            <option value="">— выберите —</option>
            {scenarios.map(s => (
              <option key={s.scenario_id} value={s.scenario_id}>
                {s.title}
              </option>
            ))}
          </select>
          {scenarios.length === 0 && (
            <p className="muted" style={{ fontSize: 13, marginTop: 4 }}>
              Нет опубликованных сценариев.
            </p>
          )}
        </label>

        <label className="field">
          <span>Первый ученик (learner)</span>
          <select value={learnerA} onChange={e => setLearnerA(e.target.value)} required>
            <option value="">— выберите —</option>
            {team.map(m => (
              <option key={m.user_id} value={m.user_id}>{m.name}</option>
            ))}
          </select>
        </label>

        <label className="field">
          <span>Второй ученик (partner)</span>
          <select value={learnerB} onChange={e => setLearnerB(e.target.value)} required>
            <option value="">— выберите —</option>
            {team.map(m => (
              <option key={m.user_id} value={m.user_id}>{m.name}</option>
            ))}
          </select>
        </label>

        <label className="field">
          <span>Дата и время</span>
          <input
            type="datetime-local"
            value={scheduledAt}
            onChange={e => setScheduledAt(e.target.value)}
            required
          />
        </label>

        {error && <p className="error">{error}</p>}
        {success && <p style={{ color: '#0a7c2f', fontSize: 14, fontWeight: 600 }}>{success}</p>}

        <button type="submit" disabled={saving}>
          {saving ? 'Сохранение...' : 'Назначить арену'}
        </button>
      </form>

      <div className="card">
        <h2 className="section-title">Все арены</h2>

        {arenas.length === 0 && (
          <p className="muted">Пока нет назначенных арен.</p>
        )}

        {arenas.map(a => (
          <div key={a.arena_assignment_id} className="competency-item">
            <div style={{ flex: 1 }}>
              <strong>{a.scenario_title}</strong>
              <p className="muted" style={{ margin: '4px 0 0', fontSize: 12 }}>
                {formatDate(a.scheduled_at)} · {statusLabel(a.status)}
              </p>
              <p className="muted" style={{ margin: '2px 0 0', fontSize: 12 }}>
                Участников: {a.participants?.length || 0}
              </p>
            </div>
            {a.arena_session_id && a.status === 'completed' && (
              <button onClick={() => navigate(`/curator/arena/${a.arena_session_id}/feedback`)}>
                Фидбек
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}