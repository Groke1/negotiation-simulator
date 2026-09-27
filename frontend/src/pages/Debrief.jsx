import { useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getCompetencyClusters, getSemanticActions, getTrainingAnalysis, getTrainingSession } from '../api';
import { ClusterBars, EmptyState, ErrorState, LoadingState } from '../components/UI';
import { semanticActionLabel } from '../app/presentation';
import { useAuth } from '../app/AuthContext';
import { resolveActiveRole } from '../app/roles';
import { formatDelta, formatNumber } from '../app/format';
import { sortMessagesChronologically } from '../app/messages';
import PublicScenarioViewer from '../components/PublicScenarioViewer';

const messageText = (message) => {
  const content = message?.content || {};
  if (content.subject && content.body) return `${content.subject}\n\n${content.body}`;
  return content.text || content.body || content.transcript || (content.media_id ? 'Голосовое сообщение' : 'Сообщение без текста');
};
const messageTime = (value) => value ? new Date(value).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' }) : '';
const SemanticCodes = ({ codes, actionNames }) => <div className="semantic-code-list">{codes.map((code) => <span key={code}>{actionNames.get(code) || code}</span>)}</div>;

function SemanticReport({ message, actionNames }) {
  const classification = message.classification;
  const annotations = Array.isArray(message.semantic_annotations) ? message.semantic_annotations : [];
  const selected = classification?.selected_semantic_action_code || message.semantic_action_code;
  const expected = classification?.expected_semantic_action_code;
  if (!selected && !annotations.length) return null;
  return <div className="message-semantic-report">
    {selected && <p><span>Выбранная метка:</span> <strong>{actionNames.get(selected) || selected}</strong></p>}
    {annotations.length ? <div><span>Метки фрагментов:</span><SemanticCodes codes={annotations.map((item) => item.semantic_action_code)} actionNames={actionNames} /></div> : null}
    {classification ? <p className={classification.correct ? 'feedback good' : 'feedback bad'}>{classification.correct ? '✓ Метка определена верно' : <>Не совпало. Правильная метка: <strong>{actionNames.get(expected) || expected}</strong></>}</p> : null}
  </div>;
}

function ProfileChanges({ changes = [], clusterNames }) {
  if (!changes.length) return <EmptyState title="Изменений профиля нет" />;
  return <div className="profile-change-list">{changes.map((change) => <article className="profile-change-row" key={change.cluster_id}><span>{clusterNames.get(change.cluster_id) || 'Компетенция'}</span><div><strong>{formatNumber(change.before)}</strong><i>→</i><strong>{formatNumber(change.after)}</strong><small className={Number(change.delta) > 0 ? 'positive' : Number(change.delta) < 0 ? 'negative' : ''}>{formatDelta(change.delta)}</small></div></article>)}</div>;
}

export default function Debrief() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const { me } = useAuth();
  const isCurator = resolveActiveRole(me) === 'curator';
  const analysis = useQuery({ queryKey: ['training-analysis', sessionId], queryFn: ({ signal }) => getTrainingAnalysis(sessionId, signal) });
  const session = useQuery({ queryKey: ['training-session', sessionId], queryFn: ({ signal }) => getTrainingSession(sessionId, signal) });
  const dictionary = useQuery({ queryKey: ['competency-clusters'], queryFn: ({ signal }) => getCompetencyClusters(signal), staleTime: 60 * 60 * 1000 });
  const actions = useQuery({ queryKey: ['semantic-actions'], queryFn: ({ signal }) => getSemanticActions(signal), staleTime: 60 * 60 * 1000 });
  const sources = [analysis, session, dictionary, actions];
  const actionNames = useMemo(() => new Map((actions.data || []).map((action) => [action.code, semanticActionLabel(action)])), [actions.data]);
  if (sources.some((source) => source.isLoading)) return <LoadingState />;
  const failed = sources.find((source) => source.error);
  if (failed) return <ErrorState error={failed.error} retry={() => sources.forEach((source) => source.refetch())} />;
  const data = analysis.data;
  const result = data.result || {};
  const clusterNames = new Map((dictionary.data || []).map((cluster) => [cluster.id, cluster.name]));
  return <div className="page training-report">
    <header className="page-header"><div><span className="eyebrow">Разбор тренировки</span><h1>{formatNumber(data.score || 0)} / 100</h1><p>{result.summary || 'Итог тренировки сформирован.'}</p></div><div className="header-actions"><PublicScenarioViewer scenario={session.data?.scenario} preparation={session.data?.preparation} currentBlock={session.data?.current_block} /><button type="button" onClick={() => navigate(isCurator ? '/curator/team' : '/learner/training')}>{isCurator ? 'К команде' : 'К тренировкам'}</button></div></header>
    {(result.goals_met?.length || result.goals_missed?.length) ? <section className="card report-outcome"><h2>Результат переговоров</h2><div className="split">{result.goals_met?.length ? <div><span className="eyebrow">Достигнуто</span><ul>{result.goals_met.map((goal) => <li key={goal}>{goal}</li>)}</ul></div> : <div><span className="eyebrow">Достигнуто</span><p>Достигнутые цели не отмечены.</p></div>}{result.goals_missed?.length ? <div><span className="eyebrow">Не достигнуто</span><ul>{result.goals_missed.map((goal) => <li key={goal}>{goal}</li>)}</ul></div> : <div><span className="eyebrow">Не достигнуто</span><p>Недостигнутые цели не отмечены.</p></div>}</div></section> : null}
    <div className="split"><section className="card"><h2>Что получилось</h2>{data.strengths?.length ? data.strengths.map((item) => <p className="feedback good" key={item.title}>✓ <strong>{item.title}</strong><br />{item.description}</p>) : <p>Сильные стороны пока не сформированы.</p>}</section><section className="card"><h2>Что улучшить</h2>{data.improvements?.length ? data.improvements.map((item) => <p className="feedback bad" key={item.title}>! <strong>{item.title}</strong><br />{item.description}</p>) : <p>Рекомендаций пока нет.</p>}</section></div>
    <section className="card" style={{ marginTop: 16 }}><h2>Компетенции в попытке</h2><ClusterBars clusters={data.competencies} dictionary={dictionary.data} />{data.competencies?.some((item) => item.comment) && <div className="form-grid" style={{ marginTop: 16 }}>{data.competencies.filter((item) => item.comment).map((item) => <p className="panel" key={item.cluster_id}><strong>{clusterNames.get(item.cluster_id) || 'Компетенция'}</strong><br />{item.comment}</p>)}</div>}</section>
    <section className="card report-transcript" style={{ marginTop: 16 }}><div className="report-section-heading"><div><span className="eyebrow">Полная история</span><h2>Диалог и семантические метки</h2></div><p>Показываем только метки, которые уже были раскрыты backend после проверки.</p></div>{session.data.messages?.length ? <div className="report-messages">{sortMessagesChronologically(session.data.messages).map((message) => <article className={`bubble ${message.author === 'learner' ? 'mine' : ''}`} key={message.message_id}><small>{message.author === 'learner' ? 'Вы' : 'Собеседник'}</small><div className="report-message-text">{messageText(message)}</div><time className="message-time">{messageTime(message.created_at)}</time><SemanticReport message={message} actionNames={actionNames} /></article>)}</div> : <EmptyState title="Сообщения не найдены" />}</section>
    <section className="card" style={{ marginTop: 16 }}><h2>Как развивался диалог</h2>{data.process?.length ? <ol className="report-process">{data.process.map((item) => <li className="panel" key={item.order}><strong>Шаг {item.order}</strong><p>{item.summary}</p>{item.semantic_action_codes?.length ? <SemanticCodes codes={item.semantic_action_codes} actionNames={actionNames} /> : null}</li>)}</ol> : <EmptyState title="Ход диалога пока не сформирован" />}</section>
    <section className="card profile-change-card" style={{ marginTop: 16 }}><h2>Изменение профиля</h2><p>Показано состояние компетенций до и после этой тренировки.</p><ProfileChanges changes={data.profile_change?.changes || []} clusterNames={clusterNames} /></section>
  </div>;
}
