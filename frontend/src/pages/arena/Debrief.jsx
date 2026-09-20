import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario } from '../../data/scenario';
import { getPatternLabel } from '../../data/patterns';

export default function ArenaDebrief() {
  const { scenarioId, outcome } = useParams();
  const navigate = useNavigate();
  const scenario = getScenario(scenarioId);
  const [step, setStep] = useState(1);

  if (!scenario || !scenario.outcomes[outcome]) {
    return (
      <div className="page">
        <h1>Разбор не найден</h1>
        <button onClick={() => navigate('/arena')}>На Арену</button>
      </div>
    );
  }

  const result = scenario.outcomes[outcome];
  const arenaResult = JSON.parse(localStorage.getItem('arena_result') || '{}');
  const patterns = arenaResult.patterns || [];
  const notes = arenaResult.notes || {};

  return (
    <div className="page">
      <header>
        <h1>Разбор Арены</h1>
        <p>{scenario.roles.learner} · {result.title}</p>
      </header>

      {step === 1 && (
        <div className="card">
          <h2 className="section-title">Что произошло</h2>
          <p>{result.summary}</p>

          {notes.goal && (
            <div className="goal-compare">
              <p><strong>Вы ставили цель:</strong> {notes.goal}</p>
              <p><strong>Результат:</strong> {result.title}</p>
            </div>
          )}

          <button onClick={() => setStep(2)}>Далее</button>
        </div>
      )}

      {step === 2 && (
        <div className="card">
          <h2 className="section-title">Как вы действовали</h2>
          <p>Использованные приёмы:</p>
          <div className="patterns-used">
            {patterns.length > 0 ? (
              patterns.map((p, i) => (
                <span key={i} className="pattern-chip">{getPatternLabel(p)}</span>
              ))
            ) : (
              <span className="muted">Нет данных</span>
            )}
          </div>
          <button onClick={() => setStep(3)}>Далее</button>
        </div>
      )}

      {step === 3 && (
        <div className="card">
          <h2 className="section-title">Что это значит</h2>
          <p>В Арене нет подсказок — вы действовали сами.</p>
          <ul className="clean-list">
            <li>Сравните с обычной практикой: где просели под давлением?</li>
            <li>Какие приёмы использовали — каких не хватило?</li>
          </ul>
          <section className="nav">
            <button onClick={() => navigate('/arena')}>На Арену</button>
            <button onClick={() => navigate('/profile')}>Профиль</button>
            <button onClick={() => navigate('/home')}>На главную</button>
          </section>
        </div>
      )}
    </div>
  );
}