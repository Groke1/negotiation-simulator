import { formatNumber } from '../app/format';

const levelPresentation = {
  excellent: { label: 'Сильная реплика', icon: '✦' },
  strong: { label: 'Уверенный ход', icon: '✓' },
  needs_attention: { label: 'Зона улучшения', icon: '↗' },
  weak: { label: 'Зона улучшения', icon: '!' }
};

const outcomePresentation = {
  positive: { label: 'Продвинуло диалог', icon: '↑' },
  neutral: { label: 'Нейтральный этап', icon: '•' },
  negative: { label: 'Требует внимания', icon: '↓' }
};

function levelFor(item) {
  if (levelPresentation[item?.level]) return item.level;
  const score = Number(item?.score || 0);
  if (score >= 85) return 'excellent';
  if (score >= 70) return 'strong';
  if (score >= 50) return 'needs_attention';
  return 'weak';
}

const clusterLabel = (id, clusterNames) => clusterNames.get(id) || 'Компетенция';

export function SemanticCodes({ codes = [], actionNames }) {
  const uniqueCodes = [...new Set(codes.filter(Boolean))];
  if (!uniqueCodes.length) return null;
  return <div className="semantic-code-list">
    {uniqueCodes.map((code) => <span key={code}><i aria-hidden="true">◆</i>{actionNames.get(code) || code}</span>)}
  </div>;
}

function ClusterLinks({ ids = [], clusterNames }) {
  const uniqueIds = [...new Set(ids.filter(Boolean))];
  if (!uniqueIds.length) return null;
  return <div className="report-cluster-links">
    <span>Связанные компетенции</span>
    <div>{uniqueIds.map((id) => <b key={id}>{clusterLabel(id, clusterNames)}</b>)}</div>
  </div>;
}

export function MessageFeedbackPanel({ feedback, actionNames, clusterNames }) {
  if (!feedback) return null;
  const level = levelFor(feedback);
  const presentation = levelPresentation[level];
  return <aside className={`message-feedback-card ${level}`}>
    <header>
      <span><i aria-hidden="true">{presentation.icon}</i>{presentation.label}</span>
      <strong>{formatNumber(feedback.score)} / 100</strong>
    </header>
    {feedback.summary && <p>{feedback.summary}</p>}
    {feedback.semantic_action_code && <SemanticCodes codes={[feedback.semantic_action_code]} actionNames={actionNames} />}
    {feedback.strengths?.length ? <div className="message-feedback-list good"><b>Что сработало</b><ul>{feedback.strengths.map((item, index) => <li key={`${item}-${index}`}>{item}</li>)}</ul></div> : null}
    {feedback.improvements?.length ? <div className="message-feedback-list improve"><b>Что улучшить</b><ul>{feedback.improvements.map((item, index) => <li key={`${item}-${index}`}>{item}</li>)}</ul></div> : null}
    <ClusterLinks ids={feedback.related_cluster_ids} clusterNames={clusterNames} />
  </aside>;
}

function ReferenceButton({ messageIds = [], onOpenMessages }) {
  if (!messageIds.length || !onOpenMessages) return null;
  return <button className="report-reference" type="button" onClick={() => onOpenMessages(messageIds)}>
    {messageIds.length === 1 ? 'Открыть реплику' : `Открыть реплики (${messageIds.length})`} →
  </button>;
}

export function ReportHighlights({ highlights, clusterNames, onOpenMessages }) {
  const groups = [
    { key: 'strengths', title: 'Что получилось', className: 'strength' },
    { key: 'improvements', title: 'Зоны роста', className: 'improvement' }
  ];
  const hasHighlights = groups.some(({ key }) => highlights?.[key]?.length);
  if (!hasHighlights) return null;

  return <section className="report-highlights">
    {groups.map(({ key, title, className }) => <section key={key} className={`report-highlight-group ${className}`}>
      <h2>{title}</h2>
      {highlights?.[key]?.length ? highlights[key].map((item, index) => <article key={`${item.title}-${index}`} className="report-highlight">
        <span className="report-highlight-icon" aria-hidden="true">{className === 'strength' ? '✦' : '↗'}</span>
        <div><h3>{item.title}</h3><p>{item.description}</p><ClusterLinks ids={item.cluster_ids} clusterNames={clusterNames} /><ReferenceButton messageIds={item.message_ids} onOpenMessages={onOpenMessages} /></div>
      </article>) : <p className="muted">Пока не выделено.</p>}
    </section>)}
  </section>;
}

export function CompetencyCards({ competencies = [], clusterNames, actionNames, onOpenMessages }) {
  if (!competencies.length) return <p className="muted">Данных по компетенциям пока нет.</p>;
  return <div className="competency-report-grid">
    {competencies.map((competency) => {
      const level = levelFor(competency);
      const score = Math.max(0, Math.min(100, Number(competency.score) || 0));
      const presentation = levelPresentation[level];
      return <article className={`competency-report-card ${level}`} key={competency.cluster_id}>
        <header><div><span className="eyebrow">Компетенция</span><h3>{clusterLabel(competency.cluster_id, clusterNames)}</h3></div><b>{formatNumber(score)}<small>/ 100</small></b></header>
        <div className="competency-progress" aria-label={`Оценка ${formatNumber(score)} из 100`}><i><b style={{ width: `${score}%` }} /></i><span>{presentation.label}</span></div>
        {competency.summary && <p className="competency-summary">{competency.summary}</p>}
        {competency.comment && <section className="competency-explanation"><span>Почему такая оценка</span><p>{competency.comment}</p></section>}
        {competency.evidence?.length ? <section className="competency-evidence"><span>Наблюдения в диалоге</span>{competency.evidence.map((item, index) => <article key={`${item.summary}-${index}`}><p>{item.summary}</p><SemanticCodes codes={item.semantic_action_codes} actionNames={actionNames} /><ReferenceButton messageIds={item.message_ids} onOpenMessages={onOpenMessages} /></article>)}</section> : null}
        {competency.recommendation && <section className="competency-recommendation"><span>Что улучшать дальше</span><p>{competency.recommendation}</p></section>}
        {competency.evidence_weight != null && <p className="competency-confidence">Достаточность наблюдений: {formatNumber(Number(competency.evidence_weight) * 100)}%</p>}
      </article>;
    })}
  </div>;
}

export function TrainingTimeline({ timeline = [], actionNames, onOpenMessages }) {
  if (!timeline.length) return <p className="muted">Ход диалога пока не сформирован.</p>;
  return <ol className="training-timeline">
    {timeline.map((item, index) => {
      const outcome = outcomePresentation[item.outcome] || outcomePresentation.neutral;
      return <li key={`${item.order}-${index}`} className={item.outcome || 'neutral'}>
        <span className="timeline-order">{item.order || index + 1}</span>
        <article><header><div>{item.block_name && <span className="eyebrow">{item.block_name}</span>}<h3>{outcome.label}</h3></div><span className="timeline-outcome" aria-label={outcome.label}>{outcome.icon}</span></header><p>{item.summary}</p><SemanticCodes codes={item.semantic_action_codes} actionNames={actionNames} /><ReferenceButton messageIds={item.message_ids} onOpenMessages={onOpenMessages} /></article>
      </li>;
    })}
  </ol>;
}
