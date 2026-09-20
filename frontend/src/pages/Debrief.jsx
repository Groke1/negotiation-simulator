import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getScenario } from '../data/scenario';
import { getHistory } from '../data/history';
import { getPatternLabel } from '../data/patterns';

export default function Debrief() {
  const { scenarioId, outcome } = useParams();
  const navigate = useNavigate();
  const scenario = getScenario(scenarioId);
  const [step, setStep] = useState(1);

  if (!scenario || !scenario.outcomes[outcome]) {
    return (
      <div className="page">
        <h1>Разбор не найден</h1>
        <button onClick={() => navigate('/home')}>На главную</button>
      </div>
    );
  }

  const result = scenario.outcomes[outcome];
  const lastSession = getHistory()[0];
  const usedPatterns = lastSession?.patterns || [];

  const competencyMap = [
    { name: 'Диагностика и слушание', pattern: 'ASK_QUESTION' },
    { name: 'Создание ценности', pattern: 'MAKE_PROPOSAL' },
    { name: 'Торг и обмен', pattern: 'REQUEST_RECIPROCITY' },
    { name: 'Цели и границы', pattern: 'HOLD_BOUNDARY' }
  ];

  return (
    <div className="page">
      <header>
        <h1>Разбор</h1>
        <p>{scenario.roles.learner} · {result.title}</p>
      </header>

      {step === 1 && (
        <div className="card">
          <h2 className="section-title">Что произошло</h2>
          <p>{result.summary}</p>
          <ul className="clean-list">
            {result.points.map((p, i) => <li key={i}>{p}</li>)}
          </ul>
          <button onClick={() => setStep(2)}>Далее</button>
        </div>
      )}

      {step === 2 && (
        <div className="card">
          <h2 className="section-title">Как вы действовали</h2>
          <p>Использованные приёмы:</p>
          <div className="patterns-used">
            {usedPatterns.length > 0 ? (
              usedPatterns.map((p, i) => (
                <span key={i} className="pattern-chip">{getPatternLabel(p)}</span>
              ))
            ) : (
              <span className="muted">Нет данных о приёмах</span>
            )}
          </div>
          <button onClick={() => setStep(3)}>Далее</button>
        </div>
      )}

      {step === 3 && (
        <div className="card">
          <h2 className="section-title">Что это говорит о компетенциях</h2>
          <div className="competency-list">
            {competencyMap.map(c => (
              <div key={c.pattern} className="competency-item">
                <span>{c.name}</span>
                <span className={usedPatterns.includes(c.pattern) ? 'mark-good' : 'mark-low'}>
                  {usedPatterns.includes(c.pattern) ? 'Проявлено' : 'Не проявлено'}
                </span>
              </div>
            ))}
          </div>
          <button onClick={() => navigate('/profile')}>Профиль</button>
          <button onClick={() => navigate('/home')}>На главную</button>
        </div>
      )}
    </div>
  );
}