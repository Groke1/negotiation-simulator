import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { generateScenario } from '../../api';

export default function CreateScenario() {
  const navigate = useNavigate();
  const [prompt, setPrompt] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  async function handleGenerate(e) {
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
        <p>Опишите тренировку своими словами — ИИ сделает черновик.</p>
      </header>

      <form onSubmit={handleGenerate} className="card">
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

        <button type="submit" disabled={loading}>
          {loading ? 'Генерация...' : 'Создать через ИИ'}
        </button>
        <button type="button" onClick={() => navigate('/constructor')} disabled={loading}>
          Назад
        </button>
      </form>
    </div>
  );
}