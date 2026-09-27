import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { getArenaHistoryDetail, getCompetencyClusters, getSemanticActions } from '../../api';
import { useAuth } from '../../app/AuthContext';
import { EmptyState, ErrorState, LoadingState } from '../../components/UI';
import { sortMessagesChronologically } from '../../app/messages';
import { semanticActionLabel } from '../../app/presentation';
import PublicScenarioViewer from '../../components/PublicScenarioViewer';
import ProfileChanges from '../../components/ProfileChanges';

const time = (value) => value ? new Date(value).toLocaleString('ru-RU') : '';
const messageText = (message) => {
  const content = message?.content || {};
  if (content.subject && content.body) return `${content.subject}\n\n${content.body}`;
  return content.text || content.body || content.transcript || (content.media_id ? 'Голосовое сообщение' : message?.text || '');
};

export default function ArenaHistory() {
  const { sessionId } = useParams();
  const { me } = useAuth();
  const details = useQuery({ queryKey: ['arena-history', sessionId], queryFn: ({ signal }) => getArenaHistoryDetail(me.id, sessionId, signal), enabled: Boolean(me?.id) });
  const actions = useQuery({ queryKey: ['semantic-actions'], queryFn: ({ signal }) => getSemanticActions(signal), staleTime: 3600000 });
  const clusters = useQuery({ queryKey: ['competency-clusters'], queryFn: ({ signal }) => getCompetencyClusters(signal), staleTime: 3600000 });
  if (details.isLoading || actions.isLoading || clusters.isLoading) return <LoadingState />;
  if (details.error || actions.error || clusters.error) return <ErrorState error={details.error || actions.error || clusters.error} retry={() => { details.refetch(); actions.refetch(); clusters.refetch(); }} />;
  const data = details.data;
  const actionNames = new Map((actions.data || []).map((action) => [action.code, semanticActionLabel(action)]));
  const clusterNames = new Map((clusters.data || []).map((cluster) => [cluster.id, cluster.name]));
  const myFeedback = (data.feedback || []).filter((item) => item.participant_id === me?.id);
  return <div className="page"><header className="page-header"><div><span className="eyebrow">Arena</span><h1>{data.scenario?.title || data.session?.scenario_title || data.scenario_title}</h1><p>Завершённые переговоры</p></div><PublicScenarioViewer scenario={data.scenario || data.session?.scenario} label="О сценарии" /></header><div className="split"><section className="card"><h2>Диалог</h2>{data.messages?.length ? sortMessagesChronologically(data.messages).map((message) => { const codes = [message.semantic_action_code, ...(message.semantic_annotations || []).map((item) => item.semantic_action_code)].filter(Boolean); return <div className="bubble" key={message.message_id}><small>{message.sender_id === me.id ? 'Вы' : 'Собеседник'}</small><div className="message-content">{messageText(message)}</div><time className="message-time">{time(message.created_at)}</time>{codes.length ? <div className="message-semantic-report"><span>Семантическая метка:</span><div className="semantic-code-list">{[...new Set(codes)].map((code) => <span key={code}>{actionNames.get(code) || code}</span>)}</div></div> : null}</div>; }) : <p>Сообщений в этой Arena нет.</p>}</section><section className="card arena-my-feedback"><h2>Ваш feedback</h2>{myFeedback.length ? myFeedback.map((feedback) => <article className="arena-feedback-result" key={feedback.feedback_id}><span className="eyebrow">Фидбек куратора</span><p>{feedback.comment}</p><time className="message-time">{time(feedback.created_at)}</time><section className="profile-change-card"><h3>Изменение профиля</h3><p>Показано состояние компетенций до и после этой Arena.</p><ProfileChanges changes={feedback.profile_change || []} clusterNames={clusterNames} emptyTitle="Показатели профиля не изменились" /></section></article>) : <EmptyState title="Фидбек пока не добавлен" description="Когда куратор оставит feedback по вашей Arena, он появится здесь." />}</section></div></div>;
}
