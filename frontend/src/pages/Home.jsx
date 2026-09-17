import { useNavigate } from 'react-router-dom';
import { scenarios } from '../data/scenario';

export default function Home() {
  const navigate = useNavigate();

  return (
    <div className="page">
      <header>
        <h1>Обучение</h1>
        <p>Выберите тренировку, чтобы отработать навык переговоров.</p>
      </header>

      <section>
        {scenarios.map(s => (
          <div key={s.id} className="card scenario-card">
            <div className="scenario-header">
              <h3>{s.title}</h3>
              <span className="badge">{s.difficulty}</span>
            </div>
            <p className="scenario-desc">{s.description}</p>
            <div className="scenario-meta">
              <span>Роль: <strong>{s.roles.learner}</strong></span>
              <span>Партнёр: <strong>{s.roles.partner}</strong></span>
            </div>
            <button onClick={() => navigate(`/chat/${s.id}`)}>Начать</button>
          </div>
        ))}
      </section>

      <section className="nav">
        <button onClick={() => navigate('/profile')}>Профиль</button>
        <button onClick={() => navigate('/history')}>История</button>
      </section>
    </div>
  );
}