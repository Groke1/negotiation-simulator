import { useState } from 'react';
import PublicScenarioGraph from './PublicScenarioGraph';

function RoleCard({ role, mine = false, opponent = false }) {
  if (!role) return null;
  return <article className={`public-role-card ${mine ? 'mine' : ''}`}>
    <span className="eyebrow">{mine ? 'Ваша роль' : opponent ? 'Собеседник' : 'Роль'}</span>
    <h3>{role.name || 'Переговорная роль'}</h3>
    {role.description && <p>{role.description}</p>}
    <dl>
      {role.goal && <><dt>Цель</dt><dd>{role.goal}</dd></>}
      {role.interests && <><dt>Интересы</dt><dd>{role.interests}</dd></>}
      {role.constraints && <><dt>Ограничения</dt><dd>{role.constraints}</dd></>}
    </dl>
    {mine && role.private_information && <section className="role-private-info"><span>Закрытая информация вашей роли</span><p>{role.private_information}</p></section>}
  </article>;
}

/** Displays only the public scenario projection returned by training/Arena APIs. */
export default function PublicScenarioViewer({ scenario, preparation, currentBlock, label = 'О сценарии' }) {
  const [open, setOpen] = useState(false);
  const publicScenario = scenario || preparation;
  const roles = publicScenario?.roles || preparation?.roles || [];
  const blocks = publicScenario?.blocks || preparation?.blocks || [];
  const currentBlockId = currentBlock?.id || currentBlock?.block_id;
  const rolesById = new Map(roles.map((role) => [role.id, role]));

  if (!publicScenario) return null;

  return <>
    <button className="secondary scenario-viewer-trigger" type="button" onClick={() => setOpen(true)}>{label}</button>
    {open && <div className="modal-overlay" role="dialog" aria-modal="true" aria-labelledby="public-scenario-title">
      <section className="modal public-scenario-viewer">
        <button className="secondary modal-close" type="button" onClick={() => setOpen(false)}>× Закрыть</button>
        <span className="eyebrow">Сценарий переговоров</span>
        <h2 id="public-scenario-title">{publicScenario.title || 'Сценарий'}</h2>
        {publicScenario.description && <p className="scenario-viewer-description">{publicScenario.description}</p>}
        {publicScenario.expanded_context && <section className="public-scenario-context"><h3>Контекст</h3><p>{publicScenario.expanded_context}</p></section>}

        <section className="public-scenario-section">
          <div className="public-scenario-heading"><div><span className="eyebrow">Роли</span><h3>Участники переговоров</h3></div></div>
          <div className="public-role-grid">
            {preparation ? <>
              <RoleCard role={preparation.my_role || rolesById.get(preparation.my_role?.id)} mine />
              <RoleCard role={preparation.partner_role || preparation.opponent_role || roles.find((role) => role.id !== preparation.my_role?.id)} opponent />
            </> : roles.map((role) => <RoleCard key={role.id} role={role} />)}
          </div>
        </section>

        <section className="public-scenario-section">
          <div className="public-scenario-heading"><div><span className="eyebrow">Схема блоков</span><h3>Состояния сценария</h3></div><p>Стрелки показывают доступные переходы между состояниями.</p></div>
          {blocks.length ? <PublicScenarioGraph blocks={blocks} currentBlockId={currentBlockId} /> : <p className="muted">Публичная схема блоков пока не передана.</p>}
        </section>
      </section>
    </div>}
  </>;
}
