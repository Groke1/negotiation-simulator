import { useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getCompetencyClusters, getSemanticActions, getTrainingAnalysis, getTrainingSession } from '../api';
import { EmptyState, ErrorState, LoadingState } from '../components/UI';
import { semanticActionLabel } from '../app/presentation';
import { useAuth } from '../app/AuthContext';
import { resolveActiveRole } from '../app/roles';
import { formatNumber } from '../app/format';
import { sortMessagesChronologically } from '../app/messages';
import PublicScenarioViewer from '../components/PublicScenarioViewer';
import { CompetencyCards, MessageFeedbackPanel, ReportHighlights, SemanticCodes, TrainingTimeline } from '../components/TrainingAnalysis';
import ProfileChanges from '../components/ProfileChanges';

const messageText = (message) => {
  const content = message?.content || {};
  if (content.subject && content.body) return `${content.subject}\n\n${content.body}`;
  return content.text || content.body || content.transcript || (content.media_id ? 'Голосовое сообщение' : 'Сообщение без текста');
};
const messageTime = (value) => value ? new Date(value).toLocaleString('ru-RU', { dateStyle: 'medium', timeStyle: 'short' }) : '';
const ANALYSIS_NOT_READY = 'TRAINING_ANALYSIS_NOT_READY';
const TRAINING_NOT_COMPLETED = 'INVALID_TRAINING_STATE';

function AnalysisPending({ onRetry, refreshing }) {
  return <div className="page analysis-pending-page"><section className="analysis-pending-card">
    <div className="analysis-pending-orbit" aria-hidden="true"><span>✦</span></div>
    <span className="eyebrow">Разбор тренировки</span>
    <h1>Анализируем вашу тренировку</h1>
    <p>Собираем разбор реплик, семантических действий и компетенций. Обычно результат появляется через пару минут.</p>
    <p className="analysis-pending-note" aria-live="polite">Страница обновляется автоматически каждые 30 секунд.</p>
    <button type="button" onClick={onRetry} disabled={refreshing}>{refreshing ? 'Проверяем…' : 'Проверить сейчас'}</button>
  </section></div>;
}

function TrainingNotCompleted({ isCurator, onBack }) {
  return <div className="page"><EmptyState title="Тренировка ещё не завершена" description="Разбор станет доступен после автоматического завершения тренировки." action={<button type="button" onClick={onBack}>{isCurator ? 'К истории команды' : 'К тренировкам'}</button>} /></div>;
}

function SemanticReport({ message, actionNames }) {
  const classification = message.classification;
  const annotations = Array.isArray(message.semantic_annotations) ? message.semantic_annotations : [];
  const selected = classification?.selected_semantic_action_code || message.semantic_action_code;
  const expected = classification?.expected_semantic_action_code;
  const isPartnerClassification = message.author === 'virtual_partner' && Boolean(classification);
  if (!selected && !annotations.length) return null;
  return <div className="message-semantic-report">
    {selected && <p><span>{isPartnerClassification ? 'Ваш вариант действия собеседника:' : 'Ваш вариант:'}</span> <strong>{actionNames.get(selected) || selected}</strong></p>}
    {annotations.length ? <div><span>Метки фрагментов:</span><SemanticCodes codes={annotations.map((item) => item.semantic_action_code)} actionNames={actionNames} /></div> : null}
    {classification ? <p className={classification.correct ? 'feedback good' : 'feedback bad'}>{classification.correct ? '✓ Метка определена верно' : <>Не совпало. Правильная метка: <strong>{actionNames.get(expected) || expected}</strong></>}</p> : null}
  </div>;
}

export default function Debrief() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const { me } = useAuth();
  const isCurator = resolveActiveRole(me) === 'curator';
  const analysis = useQuery({ queryKey: ['training-analysis', sessionId], queryFn: ({ signal }) => getTrainingAnalysis(sessionId, signal), retry: false, refetchInterval: (query) => query.state.error?.code === ANALYSIS_NOT_READY ? 30000 : false });
  const session = useQuery({ queryKey: ['training-session', sessionId], queryFn: ({ signal }) => getTrainingSession(sessionId, signal) });
  const dictionary = useQuery({ queryKey: ['competency-clusters'], queryFn: ({ signal }) => getCompetencyClusters(signal), staleTime: 60 * 60 * 1000 });
  const actions = useQuery({ queryKey: ['semantic-actions'], queryFn: ({ signal }) => getSemanticActions(signal), staleTime: 60 * 60 * 1000 });
  const sources = [analysis, session, dictionary, actions];
  const actionNames = useMemo(() => new Map((actions.data || []).map((action) => [action.code, semanticActionLabel(action)])), [actions.data]);
  if (sources.some((source) => source.isLoading)) return <LoadingState />;
  if (analysis.error?.code === ANALYSIS_NOT_READY) return <AnalysisPending onRetry={() => analysis.refetch()} refreshing={analysis.isFetching} />;
  if (analysis.error?.code === TRAINING_NOT_COMPLETED || (session.data && session.data.status !== 'completed')) return <TrainingNotCompleted isCurator={isCurator} onBack={() => navigate(isCurator ? '/curator/history' : '/learner/training')} />;
  const failed = sources.find((source) => source.error);
  if (failed) return <ErrorState error={failed.error} retry={() => sources.forEach((source) => source.refetch())} />;
  const data = analysis.data;
  const result = data.result || {};
  const clusterNames = new Map((dictionary.data || []).map((cluster) => [cluster.id, cluster.name]));
  const feedbackByMessageId = new Map((data.message_feedback || []).map((feedback) => [feedback.message_id, feedback]));
  const openMessages = (messageIds) => {
    const targetId = messageIds.find((id) => document.getElementById(`report-message-${id}`));
    document.getElementById(`report-message-${targetId}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  };
  const highlights = data.highlights || { strengths: data.strengths || [], improvements: data.improvements || [] };
  const timeline = data.timeline?.length ? data.timeline : (data.process || []).map((item) => ({ ...item, message_ids: [], outcome: 'neutral' }));
  return <div className="page training-report">
    <header className="page-header"><div><span className="eyebrow">Разбор тренировки</span><h1>{formatNumber(data.score || 0)} / 100</h1><p>{result.summary || 'Итог тренировки сформирован.'}</p></div><div className="header-actions"><PublicScenarioViewer scenario={session.data?.scenario} preparation={session.data?.preparation} currentBlock={session.data?.current_block} /><button type="button" onClick={() => navigate(isCurator ? '/curator/team' : '/learner/training')}>{isCurator ? 'К команде' : 'К тренировкам'}</button></div></header>
    {(result.goals_met?.length || result.goals_missed?.length) ? <section className="card report-outcome"><h2>Результат переговоров</h2><div className="split">{result.goals_met?.length ? <div><span className="eyebrow">Достигнуто</span><ul>{result.goals_met.map((goal) => <li key={goal}>{goal}</li>)}</ul></div> : <div><span className="eyebrow">Достигнуто</span><p>Достигнутые цели не отмечены.</p></div>}{result.goals_missed?.length ? <div><span className="eyebrow">Не достигнуто</span><ul>{result.goals_missed.map((goal) => <li key={goal}>{goal}</li>)}</ul></div> : <div><span className="eyebrow">Не достигнуто</span><p>Недостигнутые цели не отмечены.</p></div>}</div></section> : null}
    <ReportHighlights highlights={highlights} clusterNames={clusterNames} onOpenMessages={openMessages} />
    <section className="card report-competencies"><div className="report-section-heading"><div><span className="eyebrow">Компетенции в попытке</span><h2>Что показала эта тренировка</h2></div><p>Оценка относится к этой попытке, а не к текущему профилю.</p></div><CompetencyCards competencies={data.competencies} clusterNames={clusterNames} actionNames={actionNames} onOpenMessages={openMessages} /></section>
    <section className="card report-transcript" style={{ marginTop: 16 }}><div className="report-section-heading"><div><span className="eyebrow">Полная история</span><h2>Диалог и семантические метки</h2></div><p>Зелёным выделены сильные реплики, жёлтым — реплики с потенциалом для улучшения.</p></div>{session.data.messages?.length ? <div className="report-messages">{sortMessagesChronologically(session.data.messages).map((message) => { const feedback = feedbackByMessageId.get(message.message_id); return <article id={`report-message-${message.message_id}`} className={`bubble ${message.author === 'learner' ? 'mine' : ''} ${feedback ? `with-feedback ${feedback.level || ''}` : ''}`} key={message.message_id}><small>{message.author === 'learner' ? 'Вы' : 'Собеседник'}</small><div className="report-message-text">{messageText(message)}</div><time className="message-time">{messageTime(message.created_at)}</time><SemanticReport message={message} actionNames={actionNames} /><MessageFeedbackPanel feedback={feedback} actionNames={actionNames} clusterNames={clusterNames} /></article>; })}</div> : <EmptyState title="Сообщения не найдены" />}</section>
    <section className="card report-timeline" style={{ marginTop: 16 }}><div className="report-section-heading"><div><span className="eyebrow">Ключевые этапы</span><h2>Как развивался диалог</h2></div><p>Сводка опирается на реплики и переходы сценария.</p></div><TrainingTimeline timeline={timeline} actionNames={actionNames} onOpenMessages={openMessages} /></section>
    <section className="card profile-change-card" style={{ marginTop: 16 }}><h2>Изменение профиля</h2><p>Показано состояние компетенций до и после этой тренировки.</p><ProfileChanges changes={data.profile_change?.changes || []} clusterNames={clusterNames} /></section>
  </div>;
}
