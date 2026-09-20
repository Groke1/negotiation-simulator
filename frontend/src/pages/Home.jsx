import { useNavigate } from 'react-router-dom';
import { logout } from '../api';

export default function Home() {
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate('/');
  }

  return (
    <div className="page">
      <header>
        <h1>Обучение</h1>
        <p>Добрый день!</p>
      </header>

      <section className="card">
        <h2>Рекомендуем</h2>
        <h3>Тренировка: Продление контракта</h3>
        <p>Сложность: L2</p>
        <button onClick={() => navigate('/chat/sc-042')}>Начать тренировку</button>
      </section>

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