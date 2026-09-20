import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { logout, getRecommendation, getUserId, getMe } from '../api';

export default function Home() {
  const navigate = useNavigate();
  const [recommendation, setRecommendation] = useState(null);
  const [me, setMe] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const meData = await getMe();
        setMe(meData);

        try {
          const rec = await getRecommendation(meData.id);
          setRecommendation(rec);
        } catch {
          // рекомендация может не работать
        }
      } catch {}

      setLoading(false);
    }
    load();
  }, []);

  function handleLogout() {
    logout();
    navigate('/');
  }

  return (
    <div className="page">
      <header>
        <h1>Обучение</h1>
        <p>Добрый день{me?.username ? `, ${me.username}` : ''}!</p>
      </header>

      {recommendation ? (
        <section className="card">
          <h2>Рекомендуем</h2>
          <h3>{recommendation.scenario_title}</h3>
          {recommendation.reason?.text && (
            <p className="muted">{recommendation.reason.text}</p>
          )}
          {recommendation.source === 'curator_assignment' && (
            <p className="muted" style={{ fontSize: 13 }}>
              Назначено куратором
            </p>
          )}
          <button onClick={() => navigate(`/chat/${recommendation.scenario_id}`)}>
            Начать тренировку
          </button>
        </section>
      ) : (
        <section className="card">
          <h2>Рекомендуем</h2>
          <h3>Тренировка: Продление контракта</h3>
          <p>Сложность: L2</p>
          <button onClick={() => navigate('/chat/sc-042')}>Начать тренировку</button>
        </section>
      )}

      <section className="arena-cta" onClick={() => navigate('/arena')}>
        <div>
          <h2>Арена переговоров</h2>
          <p>Проверьте навык без подсказок. Сложнее, чем в Обучении.</p>
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