import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { getArenaHistoryDetail, getSemanticActions } from '../../api';
import { useAuth } from '../../app/AuthContext';
import { ErrorState, LoadingState } from '../../components/UI';
import { sortMessagesChronologically } from '../../app/messages';
import { semanticActionLabel } from '../../app/presentation';
import PublicScenarioViewer from '../../components/PublicScenarioViewer';

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
  if (details.isLoading || actions.isLoading) return <LoadingState />;
  if (details.error || actions.error) return <ErrorState error={details.error || actions.error} retry={() => { details.refetch(); actions.refetch(); }} />;
  const data = details.data;
  const actionNames = new Map((actions.data || []).map((action) => [action.code, semanticActionLabel(action)]));
  return <div className="page"><header className="page-header"><div><span className="eyebrow">Arena</span><h1>{data.scenario?.title || data.session?.scenario_title || data.scenario_title}</h1><p>Завершённые переговоры</p></div><PublicScenarioViewer scenario={data.scenario || data.session?.scenario} label="О сценарии" /></header><div className="split"><section className="card"><h2>Диалог</h2>{data.messages?.length ? sortMessagesChronologically(data.messages).map((message) => { const codes = [message.semantic_action_code, ...(message.semantic_annotations || []).map((item) => item.semantic_action_code)].filter(Boolean); return <div className="bubble" key={message.message_id}><small>{message.sender_id === me.id ? 'Вы' : 'Собеседник'}</small><div className="message-content">{messageText(message)}</div><time className="message-time">{time(message.created_at)}</time>{codes.length ? <div className="message-semantic-report"><span>Семантическая метка:</span><div className="semantic-code-list">{[...new Set(codes)].map((code) => <span key={code}>{actionNames.get(code) || code}</span>)}</div></div> : null}</div>; }) : <p>Сообщений в этой Arena нет.</p>}</section><section className="card"><h2>Фидбек куратора</h2>{data.feedback?.length ? data.feedback.map((feedback) => <article className="panel" key={feedback.feedback_id}><p>{feedback.comment}</p><time className="message-time">{time(feedback.created_at)}</time></article>) : <p>Фидбек куратора пока не добавлен.</p>}</section></div></div>;
}
