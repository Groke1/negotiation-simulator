import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { mockScenario } from '../data/mockScenario';

export default function Debrief() {
  const { scenarioId, outcome } = useParams();
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [scenario, setScenario] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Берём сценарий из localStorage или из мока
    const saved = localStorage.getItem('scenario_' + scenarioId);
    if (saved) {
      setScenario(JSON.parse(saved));
    } else {
      setScenario({ ...mockScenario, scenario_id: scenarioId });
    }
    setLoading(false);
  }, [scenarioId]);

  if (loading) return <div className="page">Загрузка...</div>;
  if (!scenario) return <div className="page">Разбор не найден</div>;

  const finalBlock = scenario.blocks?.find(b => b.id === outcome);

  if (!finalBlock) {
    return (
      <div className="page">
        <h1>Разбор не найден</h1>
        <p className="muted">Финал не найден в сценарии.</p>
        <button onClick={() => navigate('/home')}>На главную</button>
      </div>
    );
  }

  const isSuccess = finalBlock.name.toLowerCase().includes('закрыт')
    || finalBlock.name.toLowerCase().includes('успех')
    || finalBlock.name.toLowerCase().includes('договор');

  return (
    <div className="page">
      <header>
        <h1>Разбор</h1>
        <p>{scenario.title} · {finalBlock.name}</p>
      </header>

      {step === 1 && (
        <div className="card">
          <h2 className="section-title">Что произошло</h2>

          <div style={{
            padding: 16,
            background: isSuccess ? '#e3efea' : '#f7e4e1',
            borderRadius: 10,
            marginBottom: 16
          }}>
            <h3 style={{ marginTop: 0 }}>
              {isSuccess ? '✅ ' : '❌ '}
              {finalBlock.name}
            </h3>
            <p style={{ margin: 0 }}>{finalBlock.description}</p>
          </div>

          <p><strong>Сценарий:</strong> {scenario.title}</p>
          <p><strong>Роль:</strong> {scenario.roles?.[0]?.name || '—'}</p>
          <p><strong>Партнёр:</strong> {scenario.roles?.[1]?.name || '—'}</p>

          <button onClick={() => setStep(2)}>Далее</button>
        </div>
      )}

      {step === 2 && (
        <div className="card">
          <h2 className="section-title">Как вы действовали</h2>

          <p>Вы дошли до финала <strong>«{finalBlock.name}»</strong>, выбрав этот путь в диалоге.</p>

          <p className="muted">Полный разбор ходов будет доступен после подключения бэка.</p>

          <button onClick={() => setStep(3)}>Далее</button>
        </div>
      )}

      {step === 3 && (
        <div className="card">
          <h2 className="section-title">Что это говорит о компетенциях</h2>

          <div className="competency-list">
            <div className="competency-item">
              <span>Цели и границы</span>
              <span className={isSuccess ? 'mark-good' : 'mark-low'}>
                {isSuccess ? 'Проявлено' : 'Не проявлено'}
              </span>
            </div>
            <div className="competency-item">
              <span>Диагностика и слушание</span>
              <span className={isSuccess ? 'mark-good' : 'mark-low'}>
                {isSuccess ? 'Проявлено' : 'Не проявлено'}
              </span>
            </div>
            <div className="competency-item">
              <span>Торг и обмен</span>
              <span className={isSuccess ? 'mark-good' : 'mark-low'}>
                {isSuccess ? 'Проявлено' : 'Не проявлено'}
              </span>
            </div>
          </div>

          <button onClick={() => navigate('/profile')}>Профиль</button>
          <button onClick={() => navigate('/home')}>На главную</button>
        </div>
      )}
    </div>
  );
}