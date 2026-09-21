import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { logout, getMe } from '../api';
import { mockScenario } from '../data/mockScenario';

export default function Home() {
  const navigate = useNavigate();
  const [scenarios, setScenarios] = useState([]);
  const [me, setMe] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const meData = await getMe();
        setMe(meData);
      } catch {}

      // Берём сценарии из localStorage
      const published = JSON.parse(localStorage.getItem('published_scenarios') || '{}');
      const list = Object.values(published);

      // Если ничего не опубликовано — показываем мок
      if (list.length === 0) {
        setScenarios([{ ...mockScenario, scenario_id: 'sc-mock-001', status: 'published', isMock: true }]);
      } else {
        setScenarios(list);
      }

      setLoading(false);
    }
    load();
  }, []);

  function handleLogout() {
    logout();
    navigate('/');
  }

  if (loading) return <div className="page">Загрузка...</div>;

  return (
    <div className="page">
      <header>
        <h1>Обучение</h1>
        <p>Добрый день{me?.username ? `, ${me.username}` : ''}!</p>
      </header>

      <section className="card">
        <h2>Доступные тренировки</h2>

        {scenarios.length === 0 && (
          <p className="muted">Нет доступных сценариев. Методист должен опубликовать.</p>
        )}

        {scenarios.map(s => (
          <div key={s.scenario_id} style={{
            padding: '16px 0',
            borderBottom: '1px solid #f0f0f0'
          }}>
            <h3 style={{ margin: '0 0 6px' }}>{s.title}</h3>
            <p className="muted" style={{ margin: '0 0 6px', fontSize: 13 }}>
              {s.description}
            </p>
            <p className="muted" style={{ margin: '0 0 10px', fontSize: 12 }}>
              Сложность: {s.difficulty} · Блоков: {s.blocks?.length || 0}
              {s.isMock && ' · демо'}
            </p>
            <button onClick={() => navigate(`/chat/${s.scenario_id}`)}>
              Начать тренировку
            </button>
          </div>
        ))}
      </section>

      <section className="arena-cta" onClick={() => navigate('/arena')}>
        <div>
          <h2>Арена переговоров</h2>
          <p>Проверьте навык без подсказок.</p>
        </div>
        <span className="arena-arrow">→</span>
      </section>

      <section className="nav">
        <button onClick={() => navigate('/profile')}>Профиль</button>
        <button onClick={() => navigate('/history')}>История</button>
        <button onClick={handleLogout}>Выйти</button>
      </section>
    </div>
  );
}