import { formatDelta, formatNumber } from '../app/format';
import { EmptyState } from './UI';

export default function ProfileChanges({ changes = [], clusterNames = new Map(), emptyTitle = 'Изменений профиля нет' }) {
  if (!changes?.length) return <EmptyState title={emptyTitle} />;
  return <div className="profile-change-list">
    {changes.map((change) => <article className="profile-change-row" key={change.cluster_id}>
      <span>{clusterNames.get(change.cluster_id) || 'Компетенция'}</span>
      <div><strong>{formatNumber(change.before)}</strong><i>→</i><strong>{formatNumber(change.after)}</strong><small className={Number(change.delta) > 0 ? 'positive' : Number(change.delta) < 0 ? 'negative' : ''}>{formatDelta(change.delta)}</small></div>
    </article>)}
  </div>;
}
