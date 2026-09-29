import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { getArenaHistoryDetail, getCompetencyClusters, getOrganizationUsers, getSemanticActions } from '../../api';
import { useAuth } from '../../app/AuthContext';
import { resolveActiveRole } from '../../app/roles';
import { EmptyState, ErrorState, LoadingState } from '../../components/UI';
import { sortMessagesChronologically } from '../../app/messages';
import { semanticActionLabel } from '../../app/presentation';
import PublicScenarioViewer from '../../components/PublicScenarioViewer';
import ProfileChanges from '../../components/ProfileChanges';
import ArenaMessageSemanticReport from '../../components/ArenaMessageSemanticReport';

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
  const organizationUsers = useQuery({ queryKey: ['organization-users', 'arena-history'], queryFn: ({ signal }) => getOrganizationUsers({}, signal), staleTime: 5 * 60 * 1000, enabled: resolveActiveRole(me) === 'curator' });
  if (details.isLoading || actions.isLoading || clusters.isLoading) return <LoadingState />;
  if (details.error || actions.error || clusters.error) return <ErrorState error={details.error || actions.error || clusters.error} retry={() => { details.refetch(); actions.refetch(); clusters.refetch(); }} />;
  const data = details.data;
  const actionNames = new Map((actions.data || []).map((action) => [action.code, semanticActionLabel(action)]));
  const clusterNames = new Map((clusters.data || []).map((cluster) => [cluster.id, cluster.name]));
  const participantNames = new Map([...(organizationUsers.data || []).map((user) => [user.id, user.username || user.name || user.email]), ...(me?.id ? [[me.id, me.username || me.email || 'Вы']] : [])]);
  const myFeedback = (data.feedback || []).filter((item) => item.participant_id === me?.id);
  const completedBy = data.session?.completed_by || data.completed_by;
  return <div className="page"><header className="page-header"><div><span className="eyebrow">Arena</span><h1>{data.scenario?.title || data.session?.scenario_title || data.scenario_title}</h1><p>Завершённые переговоры</p></div><PublicScenarioViewer scenario={data.scenario || data.session?.scenario} label="О сценарии" /></header><div className="split"><section className="card"><h2>Диалог</h2>{data.messages?.length ? sortMessagesChronologically(data.messages).map((message) => { const mine = message.sender_id === me.id; return <div className={`bubble ${mine ? 'mine' : ''}`} key={message.message_id}><small>{participantNames.get(message.sender_id) || (mine ? (me?.username || me?.email || 'Вы') : 'Собеседник')}</small><div className="message-content">{messageText(message)}</div><time className="message-time">{time(message.created_at)}</time><ArenaMessageSemanticReport message={message} mine={mine} actionNames={actionNames} /></div>; }) : <p>Сообщений в этой Arena нет.</p>}{completedBy && <p className="system-chat-event">{completedBy === me?.id ? 'Вы завершили Arena.' : `Arena завершил ${participantNames.get(completedBy) || 'другой участник'}.`}</p>}</section><section className="card arena-my-feedback"><h2>Feedback · {me?.username || 'вы'}</h2>{myFeedback.length ? myFeedback.map((feedback) => <article className="arena-feedback-result" key={feedback.feedback_id}><span className="eyebrow">Фидбек куратора</span><p>{feedback.comment}</p><time className="message-time">{time(feedback.created_at)}</time><section className="profile-change-card"><h3>Изменение профиля</h3><p>Показано состояние компетенций до и после этой Arena.</p><ProfileChanges changes={feedback.profile_change || []} clusterNames={clusterNames} emptyTitle="Показатели профиля не изменились" /></section></article>) : <EmptyState title="Фидбек пока не добавлен" description="Когда куратор оставит feedback по вашей Arena, он появится здесь." />}</section></div></div>;
}
