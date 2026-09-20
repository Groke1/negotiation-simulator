import { useNavigate } from 'react-router-dom';
import { scenarios } from '../../data/scenario';

export default function ArenaIndex() {
  const navigate = useNavigate();

  return (
    <div className="page">
      <header>
        <h1>Арена переговоров</h1>
        <p>Без подсказок. Сложнее, чем в Обучении.</p>
      </header>

      <div className="arena-hero">
        <div className="arena-hero-icon">⚔</div>
        <h2>Что вас ждёт</h2>
        <ul className="arena-hero-list">
          <li>Без кнопок паттернов — только свободный ввод</li>
          <li>Партнёр сразу давит — без мягкого вступления</li>
          <li>Разбор с акцентом на перенос навыка</li>
        </ul>
      </div>

      <div className="card">
        <h2 className="section-title">Выберите сценарий</h2>
        {scenarios.map(s => (
          <div key={s.id} className="scenario-card">
            <div className="scenario-header">
              <h3>{s.title}</h3>
              <span className="badge">{s.difficulty}</span>
            </div>
            <p className="scenario-desc">{s.description}</p>
            <div className="scenario-meta">
              <span>Роль: <strong>{s.roles.learner}</strong></span>
              <span>Партнёр: <strong>{s.roles.partner}</strong></span>
            </div>
            <button className="arena-btn" onClick={() => navigate(`/arena/prepare/${s.id}`)}>
              Начать подготовку
            </button>
          </div>
        ))}
      </div>

      <section className="nav">
        <button onClick={() => navigate('/home')}>Обучение</button>
        <button onClick={() => navigate('/profile')}>Профиль</button>
      </section>
    </div>
  );
}