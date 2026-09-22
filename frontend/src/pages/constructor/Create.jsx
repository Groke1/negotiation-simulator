import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createScenario, generateScenario } from '../../api';

export default function CreateScenario() {
  const navigate = useNavigate();
  const [mode, setMode] = useState('manual'); // 'manual' | 'ai'
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [prompt, setPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function handleManual(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const created = await createScenario({ title, description });
      navigate(`/constructor/scenario/${created.scenario_id}`);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleAI(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const data = await generateScenario(prompt);
      navigate(`/constructor/scenario/${data.scenario_draft.scenario_id}`);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page">
      <header>
        <h1>Создание сценария</h1>
        <p>Заполните вручную или опишите словами — ИИ сделает черновик.</p>
      </header>

      <div className="nav" style={{ marginBottom: 16 }}>
        <button onClick={() => navigate('/constructor')}>← Назад</button>
      </div>

      <div style={{ display: 'flex', gap: 6, marginBottom: 16 }}>
        {[
          { id: 'manual', label: 'Вручную' },
          { id: 'ai', label: 'Через ИИ' }
        ].map(m => (
          <button
            key={m.id}
            onClick={() => { setMode(m.id); setError(''); }}
            style={{
              padding: '8px 18px',
              fontSize: 14,
              fontWeight: 700,
              borderRadius: 20,
              border: `1.5px solid ${mode === m.id ? 'var(--brand)' : 'var(--line)'}`,
              background: mode === m.id ? 'var(--brand)' : 'transparent',
              color: mode === m.id ? 'var(--paper-raised)' : 'var(--ink-soft)',
              cursor: 'pointer',
              fontFamily: 'inherit'
            }}
          >
            {m.label}
          </button>
        ))}
      </div>

      {mode === 'manual' && (
        <form onSubmit={handleManual} className="card">
          <h2 className="section-title">Новый сценарий</h2>

          <label className="field">
            <span>Название</span>
            <input
              value={title}
              onChange={e => setTitle(e.target.value)}
              placeholder="Например: Продление контракта — скидка 20%"
              required
              maxLength={200}
            />
          </label>

          <label className="field">
            <span>Описание</span>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              placeholder="Клиент требует скидку 20%, ссылаясь на конкурента. Менеджер может дать максимум 10%, но предложить расширенный SLA."
              required
              rows={5}
            />
          </label>

          {error && <p className="error">{error}</p>}

          <button type="submit" disabled={loading || !title.trim() || !description.trim()}>
            {loading ? 'Создание...' : 'Создать сценарий'}
          </button>
        </form>
      )}

      {mode === 'ai' && (
        <form onSubmit={handleAI} className="card">
          <h2 className="section-title">Черновик через ИИ</h2>

          <label className="field">
            <span>Описание тренировки</span>
            <textarea
              value={prompt}
              onChange={e => setPrompt(e.target.value)}
              required
              minLength={20}
              rows={6}
              placeholder="Например: Создай тренировку про продление контракта. Клиент требует скидку 15%, у него есть альтернативный поставщик. Менеджер может дать максимум 7%, но может предложить расширенный SLA и длинный контракт."
            />
          </label>

          {error && <p className="error">{error}</p>}

          <button type="submit" disabled={loading || prompt.length < 20}>
            {loading ? 'Генерация...' : 'Создать через ИИ'}
          </button>
        </form>
      )}
    </div>
  );
}