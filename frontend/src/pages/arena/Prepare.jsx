import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario } from '../../data/scenario';

export default function ArenaPrepare() {
  const { scenarioId } = useParams();
  const navigate = useNavigate();
  const scenario = getScenario(scenarioId);

  const [goal, setGoal] = useState('');
  const [minimum, setMinimum] = useState('');
  const [priority, setPriority] = useState('');

  function start() {
    const notes = { goal, minimum, priority };
    localStorage.setItem('arena_notes', JSON.stringify(notes));
    navigate(`/arena/chat/${scenarioId}`);
  }

  if (!scenario) return <div className="page">Сценарий не найден</div>;

  return (
    <div className="page">
      <header>
        <h1>Подготовка к Арене</h1>
        <p>{scenario.title} · {scenario.roles.learner}</p>
      </header>

      <div className="card">
        <label className="field">
          <span>Цель — чего вы хотите добиться?</span>
          <textarea
            value={goal}
            onChange={e => setGoal(e.target.value)}
            placeholder="Например: продлить контракт на 12 месяцев"
          />
        </label>

        <label className="field">
          <span>Минимально приемлемый результат</span>
          <textarea
            value={minimum}
            onChange={e => setMinimum(e.target.value)}
            placeholder="Ниже чего вы не пойдёте"
          />
        </label>

        <label className="field">
          <span>Что для вас приоритетно?</span>
          <textarea
            value={priority}
            onChange={e => setPriority(e.target.value)}
            placeholder="Цена, сроки, отношения"
          />
        </label>
      </div>

      <section className="nav">
        <button onClick={start}>Начать Арену</button>
        <button onClick={() => navigate('/arena')}>Назад</button>
      </section>
    </div>
  );
}