import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { getArenaHistoryDetail } from '../../api';
import { useAuth } from '../../app/AuthContext';
import { ErrorState, LoadingState } from '../../components/UI';
const time = (value) => value ? new Date(value).toLocaleString('ru-RU') : '';
export default function ArenaHistory() {
  const { sessionId } = useParams();
  const { me } = useAuth();
  const details = useQuery({ queryKey: ['arena-history', sessionId], queryFn: ({ signal }) => getArenaHistoryDetail(me.id, sessionId, signal), enabled: Boolean(me?.id) });
  if (details.isLoading) return <LoadingState />;
  if (details.error) return <ErrorState error={details.error} retry={details.refetch} />;
  const data = details.data;
  const myFeedback = (data.feedback || []).filter((f) => f.participant_id === me.id);
  return <div className="page"><header className="page-header"><div><span className="eyebrow">Arena</span><h1>{data.session?.scenario_title || data.scenario_title}</h1><p>Завершённые переговоры</p></div></header><div className="split"><section className="card"><h2>Диалог</h2>{data.messages?.length ? data.messages.map((message) => <div className="bubble" key={message.message_id}><small>{message.sender_id === me.id ? 'Вы' : 'Собеседник'}</small>{message.text}<time className="message-time">{time(message.created_at)}</time></div>) : <p>Сообщений в этой Arena нет.</p>}</section><section className="card"><h2>Фидбек куратора</h2>{myFeedback.length ? myFeedback.map((feedback) => <article className="panel" key={feedback.feedback_id}><p>{feedback.comment}</p><time className="message-time">{time(feedback.created_at)}</time></article>) : <p>Фидбек куратора пока не добавлен.</p>}</section></div></div>;
}