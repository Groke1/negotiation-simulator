import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getScenarios } from '../../api';
import { EmptyState, ErrorState, LoadingState, StatusBadge, TagList } from '../../components/UI';

const statuses = [['', 'Все'], ['draft', 'Черновики'], ['published', 'Опубликованные'], ['archived', 'Архив']];

export default function Constructor() {
  const navigate = useNavigate();
  const [scope, setScope] = useState('all');
  const [status, setStatus] = useState('');
  const [search, setSearch] = useState('');
  const scenarios = useQuery({
    queryKey: ['scenarios', 'methodist', scope, status],
    queryFn: ({ signal }) => getScenarios({ owned_by_me: scope === 'mine' || undefined, status: status || undefined, limit: 100 }, signal),
  });
  if (scenarios.isLoading) return <LoadingState />;
  if (scenarios.error) return <ErrorState error={scenarios.error} retry={scenarios.refetch} />;
  const items = scenarios.data.filter((item) => item.title.toLowerCase().includes(search.trim().toLowerCase()));
  return <div className="page scenarios-page">
    <header className="page-header"><div><span className="eyebrow">Методист</span><h1>Сценарии</h1><p>Создавайте практики, проверяйте граф и публикуйте готовые сценарии.</p></div><button type="button" onClick={() => navigate('/methodist/scenarios/new')}>Новый сценарий</button></header>
    <div className="scope-tabs" aria-label="Принадлежность сценариев"><button type="button" className={scope === 'all' ? 'active' : ''} onClick={() => setScope('all')}>Все сценарии</button><button type="button" className={scope === 'mine' ? 'active' : ''} onClick={() => setScope('mine')}>Мои сценарии</button></div>
    <div className="scenario-filter-row" aria-label="Фильтр по статусу"><span>Статус</span>{statuses.map(([value, label]) => <button type="button" key={label} className={status === value ? 'active' : ''} onClick={() => setStatus(value)}>{label}</button>)}</div>
    <label className="field scenario-search">Поиск<input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Название сценария" /></label>
    {items.length ? <div className="scenario-grid">{items.map((scenario) => <article className="card scenario-card scenario-link-card" key={scenario.scenario_id} role="link" tabIndex="0" onClick={() => navigate(`/methodist/scenarios/${scenario.scenario_id}`)} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') navigate(`/methodist/scenarios/${scenario.scenario_id}`); }}><header><h3>{scenario.title}</h3><StatusBadge value={scenario.status} /></header><p>{scenario.description || 'Описание сценария пока не заполнено.'}</p><TagList tags={scenario.tags} /><small className="eyebrow">{scenario.difficulty === 'easy' ? 'лёгкий' : scenario.difficulty === 'hard' ? 'высокая сложность' : 'средняя сложность'} · обновлён {scenario.updated_at ? new Date(scenario.updated_at).toLocaleDateString('ru-RU') : '—'}</small></article>)}</div> : <EmptyState title="Сценариев пока нет" action={<button type="button" onClick={() => navigate('/methodist/scenarios/new')}>Создать первый</button>} />}
  </div>;
}
